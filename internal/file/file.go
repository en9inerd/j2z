package file

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/en9inerd/j2z/internal/args"
	"github.com/en9inerd/j2z/internal/content"
	"github.com/en9inerd/j2z/internal/errs"
	"github.com/en9inerd/j2z/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

// rootFrontMatterKeys lists the keys Zola recognizes at the top level of
// a page's front matter.
var rootFrontMatterKeys = []string{
	"title",
	"description",
	"date",
	"updated",
	"weight",
	"slug",
	"draft",
	"render",
	"aliases",
	"authors",
	"path",
	"template",
	"in_search_index",
}

// dateFormats lists the date formats to try when parsing the "date" field.
var dateFormats = []string{
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
}

// MarkdownFile defines the interface for loading, processing, and saving
// a markdown file during conversion.
type MarkdownFile interface {
	Load() error
	ProcessFrontMatter() error
	ConvertToTOML(args *args.Args) error
	Save(args *args.Args) error
}

// JekyllMarkdownFile represents a Jekyll markdown post.
type JekyllMarkdownFile struct {
	Path        string
	Content     []byte
	FrontMatter []byte
}

func (f *JekyllMarkdownFile) Load() error {
	var err error
	f.Content, err = os.ReadFile(f.Path)
	return err
}

func (f *JekyllMarkdownFile) ProcessFrontMatter() error {
	fm, err := frontmatter.Extract(f.Content)
	if err != nil {
		return &errs.FrontMatterError{File: f.Path, Msg: "extraction failed", Err: err}
	}
	f.FrontMatter = fm
	return nil
}

func (f *JekyllMarkdownFile) ConvertToTOML(a *args.Args) error {
	var data map[string]any
	if err := yaml.Unmarshal(f.FrontMatter, &data); err != nil {
		return err
	}

	if a.Aliases {
		year, month, day, slug, err := parseJekyllFilename(path.Base(f.Path))
		if err != nil {
			return err
		}
		alias := fmt.Sprintf("%s/%s/%s/%s", year, month, day, slug)
		data["aliases"] = []string{alias}
	}

	if dateStr, ok := data["date"].(string); ok {
		t, err := parseDate(dateStr, a.Tz)
		if err != nil {
			return &errs.DateError{File: f.Path, Value: dateStr, Reason: "unrecognized format"}
		}
		data["date"] = t
	}

	// Map Jekyll's last_modified_at to Zola's updated field.
	if modifiedAt, ok := data["last_modified_at"]; ok {
		if dateStr, ok := modifiedAt.(string); ok {
			t, err := parseDate(dateStr, a.Tz)
			if err == nil {
				data["updated"] = t
			}
		}
		delete(data, "last_modified_at")
	}

	effectiveRootKeys := slices.Concat(rootFrontMatterKeys, a.ExtraRootKeys)

	extra := make(map[string]any)
	taxonomies := make(map[string]any)

	for key, value := range data {
		if slices.Contains(a.Taxonomies, key) {
			taxonomies[key] = value
			delete(data, key)
			continue
		}
		if !slices.Contains(effectiveRootKeys, key) {
			extra[key] = value
			delete(data, key)
		}
	}

	if len(extra) > 0 {
		data["extra"] = extra
	}
	if len(taxonomies) > 0 {
		data["taxonomies"] = taxonomies
	}

	tomlBytes, err := toml.Marshal(data)
	if err != nil {
		return err
	}

	f.FrontMatter = stripLeadingWhitespace(tomlBytes)
	return nil
}

func (f *JekyllMarkdownFile) Save(a *args.Args) error {
	outputFilePath, outputDirPath, err := getOutputPaths(f.Path, &a.JekyllDir, &a.ZolaDir)
	if err != nil {
		return err
	}

	combined := content.CombineFrontMatterAndContent(f.FrontMatter, f.Content)

	if a.DryRun {
		slog.Info("dry-run: would write", "path", outputFilePath, "size", len(combined))
		return nil
	}

	if a.OutputRoot != nil {
		relFile, _ := filepath.Rel(a.ZolaDir, outputFilePath)
		relDir, _ := filepath.Rel(a.ZolaDir, outputDirPath)
		if err := a.OutputRoot.MkdirAll(relDir, 0755); err != nil {
			return err
		}
		slog.Debug("writing file (sandboxed)", "path", relFile)
		return a.OutputRoot.WriteFile(relFile, []byte(combined), 0644)
	}

	if err := os.MkdirAll(outputDirPath, 0755); err != nil {
		return err
	}

	slog.Debug("writing file", "path", outputFilePath)
	return os.WriteFile(outputFilePath, []byte(combined), 0644)
}

// parseJekyllFilename extracts the date parts and slug from a Jekyll
// filename like "2024-01-21-amazing-node-red.md".
func parseJekyllFilename(name string) (year, month, day, slug string, err error) {
	stem, ok := strings.CutSuffix(name, ".md")
	if !ok {
		return "", "", "", "", &errs.FilenameError{Name: name, Msg: "not a .md file"}
	}
	if len(stem) < 11 || stem[4] != '-' || stem[7] != '-' || stem[10] != '-' {
		return "", "", "", "", &errs.FilenameError{Name: name, Msg: "expected YYYY-MM-DD-slug.md format"}
	}
	return stem[0:4], stem[5:7], stem[8:10], stem[11:], nil
}

// parseDate tries each known date format, with and without timezone.
func parseDate(dateStr string, tz *time.Location) (time.Time, error) {
	for _, format := range dateFormats {
		if t, err := time.ParseInLocation(format, dateStr, tz); err == nil {
			return t, nil
		}
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}

// stripLeadingWhitespace removes leading spaces/tabs from every line.
func stripLeadingWhitespace(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	for i, line := range lines {
		lines[i] = bytes.TrimLeft(line, " \t")
	}
	return bytes.Join(lines, []byte("\n"))
}
