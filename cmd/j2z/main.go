package main

import (
	"flag"
	"os"
	"strings"
	"sync"

	"github.com/en9inerd/j2z/internal/args"
	fl "github.com/en9inerd/j2z/internal/file"
	"github.com/en9inerd/j2z/internal/log"
	"github.com/en9inerd/j2z/internal/processor"
	"github.com/en9inerd/j2z/internal/timezone"
)

func splitFlag(flagValue string) []string {
	if flagValue == "" {
		return []string{}
	}
	return strings.Split(flagValue, ",")
}

var Version = "dev"

func main() {
	jekyllDirFlag := flag.String("jekyllDir", "", "Path to the Jekyll directory")
	zolaDirFlag := flag.String("zolaDir", "", "Path to the Zola directory")
	taxonomiesFlag := flag.String("taxonomies", "tags,categories", "Optional comma-separated list of taxonomies")
	tzNameFlag := flag.String("tz", "", "Optional timezone name")
	alisesFlag := flag.Bool("aliases", false, "Optional flag to enable aliases in the front matter")
	versionFlag := flag.Bool("version", false, "Print the version number")
	flag.Parse()

	if *versionFlag {
		log.Logger.Printf("version %s\n", Version)
		os.Exit(0)
	}

	cliArgs := args.Args{
		JekyllDir:  *jekyllDirFlag,
		ZolaDir:    *zolaDirFlag,
		Taxonomies: splitFlag(*taxonomiesFlag),
		Aliases:    *alisesFlag,
		Tz:         timezone.GetTimeZone(*tzNameFlag),
	}

	if cliArgs.JekyllDir == "" || cliArgs.ZolaDir == "" {
		log.Logger.Println("Error: Both --jekyllDir and --zolaDir must be provided")
		flag.Usage() // Show usage information
		os.Exit(1)
	}

	inputMdFiles, err := fl.GetMarkdownFiles(&cliArgs)
	if err != nil {
		log.Logger.Fatalf("Failed to get markdown files: %v", err)
	}

	var wg sync.WaitGroup
	for _, file := range inputMdFiles {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			mdFile := &fl.JekyllMarkdownFile{Path: file}
			if err := processor.ProcessMarkdownFile(mdFile, &cliArgs); err != nil {
				log.Logger.Printf("Failed to process file %s: %v", file, err)
			}
		}(file)
	}

	wg.Wait()
}
