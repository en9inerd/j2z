package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/en9inerd/j2z/internal/args"
)

func TestParseJekyllFilename(t *testing.T) {
	tests := []struct {
		name                          string
		input                         string
		wantY, wantM, wantD, wantSlug string
		wantErr                       bool
	}{
		{
			name:  "standard filename",
			input: "2024-01-21-amazing-node-red.md",
			wantY: "2024", wantM: "01", wantD: "21",
			wantSlug: "amazing-node-red",
		},
		{
			name:  "older filename",
			input: "2017-02-10-concordance.md",
			wantY: "2017", wantM: "02", wantD: "10",
			wantSlug: "concordance",
		},
		{
			name:    "not a .md file",
			input:   "2024-01-21-post.html",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "short.md",
			wantErr: true,
		},
		{
			name:    "missing separator",
			input:   "20240121-post.md",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y, m, d, slug, err := parseJekyllFilename(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if y != tt.wantY || m != tt.wantM || d != tt.wantD || slug != tt.wantSlug {
				t.Errorf("got (%s, %s, %s, %s), want (%s, %s, %s, %s)",
					y, m, d, slug, tt.wantY, tt.wantM, tt.wantD, tt.wantSlug)
			}
		})
	}
}

func TestStripDatePrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-01-21-amazing-node-red.md", "amazing-node-red.md"},
		{"2017-02-10-concordance.md", "concordance.md"},
		{"no-date-prefix.md", "no-date-prefix.md"},
		{"2024-1-1-short-date.md", "2024-1-1-short-date.md"}, // non-zero-padded → not stripped
		{"", ""},
		{"2024-01-21-", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := stripDatePrefix(tt.input); got != tt.want {
				t.Errorf("stripDatePrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsDigits(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1234", true},
		{"0000", true},
		{"12a4", false},
		{"", false},
		{"abcd", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isDigits(tt.input); got != tt.want {
				t.Errorf("isDigits(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetOutputPaths(t *testing.T) {
	jekyllDir := "/site/jekyll"
	zolaDir := "/site/zola"

	tests := []struct {
		name     string
		file     string
		wantFile string
		wantDir  string
	}{
		{
			name:     "standard post file",
			file:     "/site/jekyll/_posts/2024-01-21-hello-world.md",
			wantFile: "/site/zola/content/posts/hello-world.md",
			wantDir:  "/site/zola/content/posts",
		},
		{
			name:     "file without date prefix",
			file:     "/site/jekyll/_drafts/no-date.md",
			wantFile: "/site/zola/content/drafts/no-date.md",
			wantDir:  "/site/zola/content/drafts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotDir, err := getOutputPaths(tt.file, &jekyllDir, &zolaDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotFile != tt.wantFile {
				t.Errorf("file path: got %q, want %q", gotFile, tt.wantFile)
			}
			if gotDir != tt.wantDir {
				t.Errorf("dir path: got %q, want %q", gotDir, tt.wantDir)
			}
		})
	}
}

func TestConvertToTOML_BasicFields(t *testing.T) {
	f := &JekyllMarkdownFile{
		Path:        "/fake/2024-01-01-test.md",
		FrontMatter: []byte("title: My Title\ndate: 2024-01-01"),
	}

	a := &args.Args{
		Taxonomies: []string{"tags", "categories"},
		Tz:         time.UTC,
	}

	if err := f.ConvertToTOML(a); err != nil {
		t.Fatalf("ConvertToTOML failed: %v", err)
	}

	result := string(f.FrontMatter)
	if !strings.Contains(result, `title = "My Title"`) {
		t.Errorf("expected title in TOML, got:\n%s", result)
	}
	if !strings.Contains(result, "date = 2024-01-01T00:00:00Z") {
		t.Errorf("expected date in TOML, got:\n%s", result)
	}
}

func TestConvertToTOML_Taxonomies(t *testing.T) {
	f := &JekyllMarkdownFile{
		Path:        "/fake/2024-01-01-test.md",
		FrontMatter: []byte("title: Test\ntags:\n  - go\n  - cli"),
	}

	a := &args.Args{
		Taxonomies: []string{"tags", "categories"},
		Tz:         time.UTC,
	}

	if err := f.ConvertToTOML(a); err != nil {
		t.Fatalf("ConvertToTOML failed: %v", err)
	}

	result := string(f.FrontMatter)
	if !strings.Contains(result, "[taxonomies]") {
		t.Errorf("expected [taxonomies] section, got:\n%s", result)
	}
	if !strings.Contains(result, "tags") {
		t.Errorf("expected tags in taxonomies, got:\n%s", result)
	}
}

func TestConvertToTOML_ExtraKeys(t *testing.T) {
	f := &JekyllMarkdownFile{
		Path:        "/fake/2024-01-01-test.md",
		FrontMatter: []byte("title: Test\ncustom_key: custom_value"),
	}

	a := &args.Args{
		Taxonomies: []string{"tags"},
		Tz:         time.UTC,
	}

	if err := f.ConvertToTOML(a); err != nil {
		t.Fatalf("ConvertToTOML failed: %v", err)
	}

	result := string(f.FrontMatter)
	if !strings.Contains(result, "[extra]") {
		t.Errorf("expected [extra] section, got:\n%s", result)
	}
	if !strings.Contains(result, `custom_key = "custom_value"`) {
		t.Errorf("expected custom_key in extra, got:\n%s", result)
	}
}

func TestConvertToTOML_Aliases(t *testing.T) {
	f := &JekyllMarkdownFile{
		Path:        "/fake/2024-03-15-my-post.md",
		FrontMatter: []byte("title: Test"),
	}

	a := &args.Args{
		Taxonomies: []string{"tags"},
		Tz:         time.UTC,
		Aliases:    true,
	}

	if err := f.ConvertToTOML(a); err != nil {
		t.Fatalf("ConvertToTOML failed: %v", err)
	}

	result := string(f.FrontMatter)
	if !strings.Contains(result, "aliases") {
		t.Errorf("expected aliases in TOML, got:\n%s", result)
	}
	if !strings.Contains(result, "2024/03/15/my-post") {
		t.Errorf("expected alias path, got:\n%s", result)
	}
}

func TestConvertToTOML_LastModifiedAt(t *testing.T) {
	f := &JekyllMarkdownFile{
		Path:        "/fake/2024-01-01-test.md",
		FrontMatter: []byte("title: Test\ndate: 2024-01-01\nlast_modified_at: 2024-06-15 10:30"),
	}

	a := &args.Args{
		Taxonomies: []string{"tags"},
		Tz:         time.UTC,
	}

	if err := f.ConvertToTOML(a); err != nil {
		t.Fatalf("ConvertToTOML failed: %v", err)
	}

	result := string(f.FrontMatter)
	if !strings.Contains(result, "updated = 2024-06-15T10:30:00Z") {
		t.Errorf("expected updated field in TOML, got:\n%s", result)
	}
	if strings.Contains(result, "last_modified_at") {
		t.Error("last_modified_at should have been removed from output")
	}
}

func TestConvertToTOML_ExtraRootKeys(t *testing.T) {
	f := &JekyllMarkdownFile{
		Path:        "/fake/2024-01-01-test.md",
		FrontMatter: []byte("title: Test\nmy_custom_field: value"),
	}

	a := &args.Args{
		Taxonomies:    []string{"tags"},
		ExtraRootKeys: []string{"my_custom_field"},
		Tz:            time.UTC,
	}

	if err := f.ConvertToTOML(a); err != nil {
		t.Fatalf("ConvertToTOML failed: %v", err)
	}

	result := string(f.FrontMatter)
	// my_custom_field should be at root level, NOT under [extra]
	if strings.Contains(result, "[extra]") {
		t.Errorf("my_custom_field should be at root, not under [extra], got:\n%s", result)
	}
	if !strings.Contains(result, `my_custom_field = "value"`) {
		t.Errorf("expected my_custom_field at root, got:\n%s", result)
	}
}

func TestMarkdownFiles(t *testing.T) {
	// Create a temporary Jekyll directory structure.
	tmpDir := t.TempDir()

	postsDir := filepath.Join(tmpDir, "_posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatal(err)
	}

	draftsDir := filepath.Join(tmpDir, "_drafts")
	if err := os.MkdirAll(draftsDir, 0755); err != nil {
		t.Fatal(err)
	}

	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatal(err)
	}

	for _, f := range []string{
		filepath.Join(postsDir, "2024-01-01-post1.md"),
		filepath.Join(postsDir, "2024-01-02-post2.md"),
		filepath.Join(draftsDir, "2024-01-03-draft1.md"),
		filepath.Join(assetsDir, "image.md"), // should be skipped
	} {
		if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var files []string
	for path, err := range MarkdownFiles(tmpDir) {
		if err != nil {
			t.Fatalf("MarkdownFiles yielded error: %v", err)
		}
		files = append(files, path)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(files), files)
	}
}
