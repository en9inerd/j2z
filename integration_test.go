package main_test

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/en9inerd/j2z/internal/args"
	"github.com/en9inerd/j2z/internal/file"
	"github.com/en9inerd/j2z/internal/processor"
)

func TestIntegrationSampleData(t *testing.T) {
	jekyllDir := "sample_data/jekyll"
	expectedDir := "sample_data/zola/content/posts"

	if _, err := os.Stat(jekyllDir); os.IsNotExist(err) {
		t.Skip("sample_data not available")
	}

	outDir := t.TempDir()

	a := &args.Args{
		JekyllDir:  jekyllDir,
		ZolaDir:    outDir,
		Taxonomies: []string{"tags", "categories"},
		Tz:         mustLoadLocation(t, "America/New_York"),
	}

	var (
		wg       sync.WaitGroup
		total    atomic.Int64
		errCount atomic.Int64
		sem      = make(chan struct{}, 4)
	)

	for path, err := range file.MarkdownFiles(jekyllDir) {
		if err != nil {
			t.Fatalf("MarkdownFiles error: %v", err)
		}

		total.Add(1)
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			mdFile := &file.JekyllMarkdownFile{Path: path}
			if err := processor.ProcessMarkdownFile(mdFile, a); err != nil {
				t.Errorf("failed to process %s: %v", path, err)
				errCount.Add(1)
			}
		}()
	}
	wg.Wait()

	if total.Load() == 0 {
		t.Fatal("no input files found")
	}
	if errCount.Load() > 0 {
		t.Fatalf("%d files failed to process", errCount.Load())
	}

	expectedFiles, err := os.ReadDir(expectedDir)
	if err != nil {
		t.Fatalf("failed to read expected dir: %v", err)
	}

	outputPostsDir := filepath.Join(outDir, "content", "posts")
	for _, entry := range expectedFiles {
		if entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			expected, err := os.ReadFile(filepath.Join(expectedDir, entry.Name()))
			if err != nil {
				t.Fatalf("read expected: %v", err)
			}

			actual, err := os.ReadFile(filepath.Join(outputPostsDir, entry.Name()))
			if err != nil {
				t.Fatalf("read actual: %v", err)
			}

			if string(actual) != string(expected) {
				// Find first difference for a useful error message.
				diffLine := findFirstDiffLine(string(expected), string(actual))
				t.Errorf("output mismatch for %s (first diff near line %d)\n--- expected ---\n%s\n--- actual ---\n%s",
					entry.Name(), diffLine, truncate(string(expected), 500), truncate(string(actual), 500))
			}
		})
	}
}

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("failed to load timezone %q: %v", name, err)
	}
	return loc
}

func findFirstDiffLine(a, b string) int {
	line := 1
	i := 0
	for i < len(a) && i < len(b) {
		if a[i] != b[i] {
			return line
		}
		if a[i] == '\n' {
			line++
		}
		i++
	}
	return line
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
