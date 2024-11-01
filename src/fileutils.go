package main

import (
	"os"
	"path/filepath"
	"regexp"
)

// Get all markdown files in the Jekyll directory
func getMarkdownFiles(jekyllDir *string) ([]string, error) {
	var files []string

	dirs, err := os.ReadDir(*jekyllDir)
	if err != nil {
		return nil, err
	}

	// Walk through all directories starting with an underscore
	for _, dir := range dirs {
		if dir.IsDir() && dir.Name()[0] == '_' {
			err := filepath.Walk(filepath.Join(*jekyllDir, dir.Name()), func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if filepath.Ext(path) == ".md" {
					files = append(files, path)
				}

				return nil
			})

			if err != nil {
				return nil, err
			}
		}
	}

	return files, nil
}

func getOutputFilePath(file string, jekyllDir *string, zolaDir *string) string {
	relPath, err := filepath.Rel(*jekyllDir, file)

	if relPath[0] == '_' {
		relPath = relPath[1:]
	}

	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}-`)

	file = filepath.Base(relPath)
	relPath = filepath.Dir(relPath)

	file = re.ReplaceAllString(file, "")
	relPath = filepath.Join(relPath, file)

	if err != nil {
		logger.Println("Error getting relative path: " + file)
		os.Exit(1)
	}

	return filepath.Join(*zolaDir, "content", relPath)
}
