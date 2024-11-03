package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"
)

// struct to hold all arguments
type Args struct {
	jekyllDir  string
	zolaDir    string
	taxonomies []string
	tz         *time.Location
	aliases    bool
}

var logger *log.Logger = log.New(os.Stdout, "j2z: ", 0)

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
		logger.Printf("version %s\n", Version)
		os.Exit(0)
	}

	cliArgs := Args{
		jekyllDir:  *jekyllDirFlag,
		zolaDir:    *zolaDirFlag,
		taxonomies: splitFlag(*taxonomiesFlag),
		aliases:    *alisesFlag,
		tz:         getTimeZone(*tzNameFlag),
	}

	if cliArgs.jekyllDir == "" || cliArgs.zolaDir == "" {
		logger.Println("Error: Both --jekyllDir and --zolaDir must be provided")
		flag.Usage() // Show usage information
		os.Exit(1)
	}

	inputMdFiles, err := getMarkdownFiles(&cliArgs)
	if err != nil {
		logger.Fatalf("Failed to get markdown files: %v", err)
	}

	for _, file := range inputMdFiles {
		mdFile := &JekyllMarkdownFile{Path: file}
		if err := processMarkdownFile(mdFile, &cliArgs); err != nil {
			logger.Printf("Failed to process file %s: %v", file, err)
		}
	}
}
