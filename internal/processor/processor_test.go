package processor

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/en9inerd/j2z/internal/args"
	"github.com/en9inerd/j2z/internal/file"
)

func TestProcessMarkdownFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "_posts")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := "---\ntitle: Test Post\ndate: 2024-01-01\ntags:\n  - go\n---\n\nHello world.\n"
	inputFile := filepath.Join(inputDir, "2024-01-01-test-post.md")
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	a := &args.Args{
		JekyllDir:  tmpDir,
		ZolaDir:    filepath.Join(tmpDir, "zola_output"),
		Taxonomies: []string{"tags", "categories"},
		Tz:         time.UTC,
	}

	mdFile := &file.JekyllMarkdownFile{Path: inputFile}
	if err := ProcessMarkdownFile(mdFile, a); err != nil {
		t.Fatalf("ProcessMarkdownFile failed: %v", err)
	}

	outputFile := filepath.Join(a.ZolaDir, "content", "posts", "test-post.md")
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}

	output := string(data)
	if output[:4] != "+++\n" {
		t.Error("output should start with TOML delimiter")
	}
	if !strings.Contains(output, `title = "Test Post"`) {
		t.Error("output should contain title")
	}
	if !strings.Contains(output, "Hello world.") {
		t.Error("output should contain body")
	}
}

func TestConcurrentProcessing(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tmpDir := t.TempDir()
		inputDir := filepath.Join(tmpDir, "_posts")
		if err := os.MkdirAll(inputDir, 0755); err != nil {
			t.Fatal(err)
		}

		fileCount := 10
		for i := range fileCount {
			content := "---\ntitle: Post\ndate: 2024-01-01\n---\n\nBody.\n"
			name := filepath.Join(inputDir, "2024-01-01-post-"+strconv.Itoa(i)+".md")
			if err := os.WriteFile(name, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}

		a := &args.Args{
			JekyllDir:  tmpDir,
			ZolaDir:    filepath.Join(tmpDir, "zola_out"),
			Taxonomies: []string{"tags"},
			Tz:         time.UTC,
		}

		var files []string
		for path, err := range file.MarkdownFiles(tmpDir) {
			if err != nil {
				t.Fatal(err)
			}
			files = append(files, path)
		}

		if len(files) != fileCount {
			t.Fatalf("expected %d files, got %d", fileCount, len(files))
		}

		var (
			wg       sync.WaitGroup
			errCount atomic.Int64
			sem      = make(chan struct{}, 4)
		)

		for _, f := range files {
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()

				mdFile := &file.JekyllMarkdownFile{Path: f}
				if err := ProcessMarkdownFile(mdFile, a); err != nil {
					errCount.Add(1)
				}
			}()
		}

		wg.Wait()

		if errCount.Load() != 0 {
			t.Errorf("expected 0 errors, got %d", errCount.Load())
		}

		outputDir := filepath.Join(a.ZolaDir, "content", "posts")
		entries, err := os.ReadDir(outputDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != fileCount {
			t.Errorf("expected %d output files, got %d", fileCount, len(entries))
		}
	})
}
