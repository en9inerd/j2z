package file

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/en9inerd/j2z/internal/args"
	"github.com/en9inerd/j2z/internal/content"
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
	ConvertToTOML(args *args.Args) error
	Save(args *args.Args) error
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

func (f *JekyllMarkdownFile) ConvertToTOML(args *args.Args) error {
	var data map[string]interface{}
	err := yaml.Unmarshal(f.FrontMatter, &data)
	if err != nil {
		return err
	}

	// Add alias if the flag is set
	if args.Aliases {
		fileName := path.Base(f.Path)

		re := regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})-(.*)\.md`)
		match := re.FindStringSubmatch(fileName)

		if len(match) < 5 {
			return fmt.Errorf("invalid file name format: %s", fileName)
		}

		alias := fmt.Sprintf("%s/%s/%s/%s", match[1], match[2], match[3], match[4])
		data["aliases"] = []string{alias}
	}

	// Parse "date" field if it exists and is valid
	if dateStr, ok := data["date"].(string); ok {
		var t time.Time
		var err error

		for _, format := range dateFormats {
			if t, err = time.ParseInLocation(format, dateStr, args.Tz); err == nil {
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
		if slices.Contains(args.Taxonomies, key) {
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

func (f *JekyllMarkdownFile) Save(args *args.Args) error {
	outputFilePath, outputDirPath, err := getOutputPaths(f.Path, &args.JekyllDir, &args.ZolaDir)
	if err != nil {
		return err
	}
	combined := content.CombineFrontMatterAndContent(f.FrontMatter, f.Content)

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
