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
	var b strings.Builder
	fmt.Fprintf(&b, "j2z version %s", version)
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				if len(kv.Value) >= 7 {
					fmt.Fprintf(&b, " (%s)", kv.Value[:7])
				}
			case "vcs.time":
				fmt.Fprintf(&b, " built %s", kv.Value)
			}
		}
	}
	return b.String()
}

func main() {
	jekyllDirFlag := flag.String("jekyllDir", "", "Path to the Jekyll directory")
	zolaDirFlag := flag.String("zolaDir", "", "Path to the Zola directory")
	taxonomiesFlag := flag.String("taxonomies", "tags,categories", "Optional comma-separated list of taxonomies")
	extraKeysFlag := flag.String("extraRootKeys", "", "Optional comma-separated list of additional root front matter keys")
	tzNameFlag := flag.String("tz", "", "Optional timezone name")
	aliasesFlag := flag.Bool("aliases", false, "Enable aliases in the front matter")
	dryRunFlag := flag.Bool("dry-run", false, "Preview conversion without writing files")
	versionFlag := flag.Bool("version", false, "Print the version number")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging")
	quietFlag := flag.Bool("quiet", false, "Suppress all output except errors")
	flag.Parse()

	switch {
	case *quietFlag:
		applog.Level.Set(slog.LevelWarn)
	case *verboseFlag:
		applog.Level.Set(slog.LevelDebug)
	}

	if *versionFlag {
		fmt.Println(versionString())
		os.Exit(0)
	}

	cliArgs := args.Args{
		JekyllDir:     *jekyllDirFlag,
		ZolaDir:       *zolaDirFlag,
		Taxonomies:    splitFlag(*taxonomiesFlag),
		ExtraRootKeys: splitFlag(*extraKeysFlag),
		Aliases:       *aliasesFlag,
		DryRun:        *dryRunFlag,
		Tz:            timezone.GetTimeZone(*tzNameFlag),
	}

	if cliArgs.JekyllDir == "" || cliArgs.ZolaDir == "" {
		slog.Error("both --jekyllDir and --zolaDir must be provided")
		flag.Usage()
		os.Exit(1)
	}

	if !cliArgs.DryRun {
		if err := os.MkdirAll(cliArgs.ZolaDir, 0755); err != nil {
			slog.Error("cannot create Zola directory", "dir", cliArgs.ZolaDir, "err", err)
			os.Exit(1)
		}
		outputRoot, err := os.OpenRoot(cliArgs.ZolaDir)
		if err != nil {
			slog.Error("cannot open Zola directory", "dir", cliArgs.ZolaDir, "err", err)
			os.Exit(1)
		}
		defer outputRoot.Close()
		cliArgs.OutputRoot = outputRoot
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
	slog.Info("conversion complete", "total", t, "succeeded", t-failed, "failed", failed)

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
