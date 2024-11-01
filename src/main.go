package main

import (
	"flag"
	"log"
	"os"
)

var logger *log.Logger = log.New(os.Stdout, "j2z: ", 0)

func main() {
	jekyllDir := flag.String("jekyllDir", "", "Path to the Jekyll directory")
	zolaDir := flag.String("zolaDir", "", "Path to the Zola directory")
	tzName := flag.String("timezone", "", "Optional timezone")
	flag.Parse()

	if *jekyllDir == "" || *zolaDir == "" {
		logger.Println("Error: Both --jekyllDir and --zolaDir must be provided")
		flag.Usage() // Show usage information
		os.Exit(1)
	}

	tz := getTimeZone(*tzName)

	inputMdFiles, err := getMarkdownFiles(jekyllDir)
	if err != nil {
		logger.Fatalf("Failed to get markdown files: %v", err)
	}

	for _, file := range inputMdFiles {
		mdFile := &JekyllMarkdownFile{Path: file}
		if err := processMarkdownFile(mdFile, jekyllDir, zolaDir, tz); err != nil {
			logger.Printf("Failed to process file %s: %v", file, err)
		}
	}
}
