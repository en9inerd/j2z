package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/en9inerd/go-pkgs/flagpair"
	"github.com/en9inerd/j2z/internal/args"
	"github.com/en9inerd/j2z/internal/errs"
	"github.com/en9inerd/j2z/internal/file"
	applog "github.com/en9inerd/j2z/internal/log"
	"github.com/en9inerd/j2z/internal/processor"
	"github.com/en9inerd/j2z/internal/timezone"
)

var version = "dev"

func splitFlag(flagValue string) []string {
	if flagValue == "" {
		return []string{}
	}
	return strings.Split(flagValue, ",")
}

func versionString() string {
	var revision, buildTime string
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				if len(kv.Value) >= 7 {
					revision = kv.Value[:7]
				}
			case "vcs.time":
				buildTime = kv.Value
			}
		}
	}
	s := "j2z version " + version
	if revision != "" {
		s += " (" + revision + ")"
	}
	if buildTime != "" {
		s += " built " + buildTime
	}
	return s
}

func main() {
	r := flagpair.New("j2z")
	jekyllDir := r.String("jekyll-dir", "j", "", "Path to the Jekyll directory")
	zolaDir := r.String("zola-dir", "z", "", "Path to the Zola directory")
	taxonomies := r.String("taxonomies", "", "tags,categories", "Optional comma-separated list of taxonomies")
	extraKeys := r.String("extra-root-keys", "", "", "Optional comma-separated list of additional root front matter keys")
	tzName := r.String("tz", "", "", "Optional timezone name")
	aliases := r.Bool("aliases", "", false, "Enable aliases in the front matter")
	dryRun := r.Bool("dry-run", "", false, "Preview conversion without writing files")
	showVersion := r.Bool("version", "", false, "Print the version number")
	verbose := r.Bool("verbose", "v", false, "Enable verbose logging")
	quiet := r.Bool("quiet", "q", false, "Suppress all output except errors")

	if err := r.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		slog.Error("invalid arguments", "err", err)
		os.Exit(1)
	}

	switch {
	case *quiet:
		applog.Level.Set(slog.LevelWarn)
	case *verbose:
		applog.Level.Set(slog.LevelDebug)
	}

	if *showVersion {
		fmt.Println(versionString())
		os.Exit(0)
	}

	cliArgs := args.Args{
		JekyllDir:     *jekyllDir,
		ZolaDir:       *zolaDir,
		Taxonomies:    splitFlag(*taxonomies),
		ExtraRootKeys: splitFlag(*extraKeys),
		Aliases:       *aliases,
		DryRun:        *dryRun,
		Tz:            timezone.GetTimeZone(*tzName),
	}

	if cliArgs.JekyllDir == "" || cliArgs.ZolaDir == "" {
		slog.Error("both --jekyll-dir and --zola-dir must be provided")
		r.FlagSet().Usage()
		os.Exit(1)
	}

	var (
		wg       sync.WaitGroup
		total    atomic.Int64
		errCount atomic.Int64
		sem      = make(chan struct{}, runtime.NumCPU())
	)

	for path, err := range file.MarkdownFiles(cliArgs.JekyllDir) {
		if err != nil {
			slog.Error("error walking directory", "err", err)
			errCount.Add(1)
			continue
		}

		total.Add(1)
		wg.Add(1)
		sem <- struct{}{} // acquire
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // release

			mdFile := &file.JekyllMarkdownFile{Path: path}
			if err := processor.ProcessMarkdownFile(mdFile, &cliArgs); err != nil {
				logProcessingError(path, err)
				errCount.Add(1)
				return
			}
			slog.Info("converted", "file", path)
		}()
	}

	wg.Wait()

	t := int(total.Load())
	failed := int(errCount.Load())
	slog.Info("conversion complete", "total", t, "succeeded", max(0, t-failed), "failed", failed)

	if failed > 0 {
		os.Exit(1)
	}
}

func logProcessingError(path string, err error) {
	if fmErr, ok := errors.AsType[*errs.FrontMatterError](err); ok {
		slog.Error("front matter error", "file", fmErr.File, "msg", fmErr.Msg, "err", fmErr.Err)
	} else if fnErr, ok := errors.AsType[*errs.FilenameError](err); ok {
		slog.Error("filename error", "file", path, "name", fnErr.Name, "msg", fnErr.Msg)
	} else if dtErr, ok := errors.AsType[*errs.DateError](err); ok {
		slog.Error("date parse error", "file", dtErr.File, "value", dtErr.Value, "reason", dtErr.Reason)
	} else {
		slog.Error("failed to process file", "file", path, "err", err)
	}
}
