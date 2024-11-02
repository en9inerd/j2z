package main

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// List of root front matter keys in Zola
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

// List of date formats to try when parsing the "date" field
var dateFormats = []string{
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
}

type MarkdownFile interface {
	Load() error
	ProcessFrontMatter() error
	ConvertToTOML(args *Args) error
	Save(args *Args) error
}

type JekyllMarkdownFile struct {
	Path        string
	Content     []byte
	FrontMatter []byte
}

func (f *JekyllMarkdownFile) Load() error {
	var err error
	f.Content, err = os.ReadFile(f.Path)
	if err != nil {
		return err
	}

	return nil
}

func (f *JekyllMarkdownFile) ProcessFrontMatter() error {
	re := regexp.MustCompile(`(?s)---\n(.*?)\n---`)
	match := re.FindStringSubmatch(string(f.Content))

	if len(match) == 0 {
		return fmt.Errorf("no front matter found in file %s", f.Path)
	}

	f.FrontMatter = []byte(match[1])
	return nil
}

func (f *JekyllMarkdownFile) ConvertToTOML(args *Args) error {
	var data map[string]interface{}
	err := yaml.Unmarshal(f.FrontMatter, &data)
	if err != nil {
		return err
	}

	// Parse "date" field if it exists and is valid
	if dateStr, ok := data["date"].(string); ok {
		var t time.Time
		var err error

		for _, format := range dateFormats {
			if t, err = time.ParseInLocation(format, dateStr, args.tz); err == nil {
				data["date"] = t
				break
			}

			if t, err = time.Parse(format, dateStr); err == nil {
				data["date"] = t
				break
			}
		}

		if err != nil {
			return fmt.Errorf("invalid date format in front matter: %s", dateStr)
		}
	}

	// Separate root keys from non-root keys
	extra := make(map[string]interface{})
	taxonomies := make(map[string]interface{})
	for key, value := range data {
		// Check if key is taxonomy
		if slices.Contains(args.taxonomies, key) {
			taxonomies[key] = value
			delete(data, key) // Remove the taxonomy key from the original map
			continue
		}

		// Check if the key is not a root key
		if !slices.Contains(rootFrontMatterKeys, key) {
			extra[key] = value
			delete(data, key) // Remove the non-root key from the original map
		}
	}

	// Add the "extra" section if there are any non-root keys
	if len(extra) > 0 {
		data["extra"] = extra
	}

	// Add the "taxonomies" section if there are any taxonomies
	if len(taxonomies) > 0 {
		data["taxonomies"] = taxonomies
	}

	// Marshal the data into TOML format
	f.FrontMatter, err = toml.Marshal(data)
	if err != nil {
		return err
	}

	// Post-process the TOML data to remove identation
	re := regexp.MustCompile(`(?m)^[ \t]+`)
	f.FrontMatter = re.ReplaceAll(f.FrontMatter, []byte(""))

	return nil
}

func (f *JekyllMarkdownFile) Save(args *Args) error {
	outputFilePath, outputDirPath, err := getOutputPaths(f.Path, &args.jekyllDir, &args.zolaDir)
	if err != nil {
		return err
	}
	combined := combineFrontMatterAndContent(f.FrontMatter, f.Content)

	if _, err := os.Stat(outputDirPath); os.IsNotExist(err) {
		err := os.MkdirAll(outputDirPath, 0755)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(outputFilePath, []byte(combined), 0644)
	if err != nil {
		return err
	}

	return nil
}
