package main

import (
	"fmt"
	"os"
	"path/filepath"
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

func main() {
	args := getAndCheckArgs()

	mdFiles, err := getMarkdownFiles(args["jekyllDir"])
	if err != nil {
		fmt.Println("Error getting markdown files")
		os.Exit(1)
	}

	for _, file := range mdFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("Error reading file: " + file)
			os.Exit(1)
		}

		frontMatter := extractFrontMatter(string(content))
		if frontMatter == nil {
			fmt.Println("No front matter found in: " + file)
			continue
		}

		tomlData, err := convertFrontMatterToTOML(*frontMatter)
		if err != nil {
			fmt.Println("Error converting front matter to TOML: " + file)
		}

		fmt.Println(string(tomlData))

		_ = tomlData

		// Write the TOML front matter to the Zola directory
		// zolaFile := filepath.Join(args["zolaDir"], "content", filepath.Base(file))
		// err = os.WriteFile(zolaFile, tomlData, 0644)

		// if err != nil {
		// 	fmt.Println("Error writing file: " + zolaFile)
		// }
	}
}

// Get and check the command line arguments
func getAndCheckArgs() map[string]string {
	var jekyllDir string
	var zolaDir string

	var args = os.Args[1:]
	if len(args) != 2 {
		fmt.Println(
			"No arguments provided" + "\n\n" + "Usage: j2z <jekyll-dir> <zola-dir>",
		)
		os.Exit(1)
	} else {
		jekyllDir = args[0]
		zolaDir = args[1]
	}

	if _, err := os.Stat(jekyllDir); os.IsNotExist(err) {
		fmt.Println("Jekyll directory does not exist")
		os.Exit(1)
	}

	if _, err := os.Stat(zolaDir); os.IsNotExist(err) {
		fmt.Println("Zola directory does not exist")
		os.Exit(1)
	}

	return map[string]string{
		"jekyllDir": jekyllDir,
		"zolaDir":   zolaDir,
	}
}

// Get all markdown files in the Jekyll directory
func getMarkdownFiles(jekyllDir string) ([]string, error) {
	var files []string

	dirs, err := os.ReadDir(jekyllDir)
	if err != nil {
		return nil, err
	}

	// Walk through all directories starting with an underscore
	for _, dir := range dirs {
		if dir.IsDir() && dir.Name()[0] == '_' {
			err := filepath.Walk(filepath.Join(jekyllDir, dir.Name()), func(path string, info os.FileInfo, err error) error {
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

// Extracts the YAML front matter from a markdown file
func extractFrontMatter(content string) *string {
	re := regexp.MustCompile(`(?s)---\n(.*?)\n---`)
	match := re.FindStringSubmatch(content)

	if len(match) == 0 {
		return nil
	}

	return &match[1]
}

// Converts YAML front matter to TOML format
func convertFrontMatterToTOML(frontMatter string) ([]byte, error) {
	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(frontMatter), &data)
	if err != nil {
		return nil, err
	}

	// Parse "date" field if it exists and is valid
	if dateStr, ok := data["date"].(string); ok {
		data["date"], err = time.Parse("2006-01-02 15:04", dateStr)
		if err != nil {
			return nil, err
		}
	}

	// Separate root keys from non-root keys
	extra := make(map[string]interface{})
	for key, value := range data {
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

	// Marshal the data into TOML format
	var tomlData []byte
	tomlData, err = toml.Marshal(data)
	if err != nil {
		return nil, err
	}

	return tomlData, nil
}
