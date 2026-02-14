package file

import (
	"io/fs"
	"iter"
	"os"
	"path/filepath"
)

// MarkdownFiles returns an iterator that lazily yields markdown file paths
// found in underscore-prefixed subdirectories of the given directory.
// Processing can start before the full directory walk completes.
func MarkdownFiles(dir string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		dirs, err := os.ReadDir(dir)
		if err != nil {
			yield("", err)
			return
		}

		for _, d := range dirs {
			if !d.IsDir() || d.Name()[0] != '_' {
				continue
			}
			err := filepath.WalkDir(filepath.Join(dir, d.Name()), func(path string, e fs.DirEntry, err error) error {
				if err != nil {
					if !yield("", err) {
						return filepath.SkipAll
					}
					return nil
				}
				if filepath.Ext(path) == ".md" {
					if !yield(path, nil) {
						return filepath.SkipAll
					}
				}
				return nil
			})
			if err != nil {
				if !yield("", err) {
					return
				}
			}
		}
	}
}

func getOutputPaths(file string, jekyllDir *string, zolaDir *string) (string, string, error) {
	relPath, err := filepath.Rel(*jekyllDir, file)
	if err != nil {
		return "", "", err
	}

	if len(relPath) > 0 && relPath[0] == '_' {
		relPath = relPath[1:]
	}

	name := filepath.Base(relPath)
	dir := filepath.Dir(relPath)

	name = stripDatePrefix(name)
	relPath = filepath.Join(dir, name)

	return filepath.Join(*zolaDir, "content", relPath), filepath.Join(*zolaDir, "content", dir), nil
}

// stripDatePrefix removes a leading "YYYY-MM-DD-" prefix from a filename
// if present.
func stripDatePrefix(name string) string {
	if len(name) >= 11 &&
		isDigits(name[0:4]) && name[4] == '-' &&
		isDigits(name[5:7]) && name[7] == '-' &&
		isDigits(name[8:10]) && name[10] == '-' {
		return name[11:]
	}
	return name
}

// isDigits reports whether every byte in s is an ASCII digit.
func isDigits(s string) bool {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}
