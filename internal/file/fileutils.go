package file

import (
	"os"
	"path/filepath"
	"regexp"

	Args "github.com/en9inerd/j2z/internal/args"
)

// Get all markdown files in the Jekyll directory
func GetMarkdownFiles(args *Args.Args) ([]string, error) {
	var files []string

	dirs, err := os.ReadDir(args.JekyllDir)
	if err != nil {
		return nil, err
	}

	// Walk through all directories starting with an underscore
	for _, dir := range dirs {
		if dir.IsDir() && dir.Name()[0] == '_' {
			err := filepath.Walk(filepath.Join(args.JekyllDir, dir.Name()), func(path string, info os.FileInfo, err error) error {
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

func getOutputPaths(file string, jekyllDir *string, zolaDir *string) (string, string, error) {
	relPath, err := filepath.Rel(*jekyllDir, file)
	if err != nil {
		return "", "", err
	}

	if relPath[0] == '_' {
		relPath = relPath[1:]
	}

	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}-`)

	file = filepath.Base(relPath)
	dir := filepath.Dir(relPath)

	file = re.ReplaceAllString(file, "")
	relPath = filepath.Join(dir, file)

	return filepath.Join(*zolaDir, "content", relPath), filepath.Join(*zolaDir, "content", dir), nil
}
