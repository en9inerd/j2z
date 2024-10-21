package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

var logger *log.Logger = log.New(os.Stdout, "j2z: ", 0)

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

func main() {
	// Define the command-line flags (named arguments)
	jekyllDir := flag.String("jekyllDir", "", "Path to the Jekyll directory")
	zolaDir := flag.String("zolaDir", "", "Path to the Zola directory")
	timezone := flag.String("timezone", "", "Optional timezone (e.g., 'America/New_York'). Defaults to local timezone.")
	flag.Parse()

	// Validate required arguments
	if *jekyllDir == "" || *zolaDir == "" {
		logger.Println("Error: Both --jekyllDir and --zolaDir must be provided")
		flag.Usage() // Show usage information
		os.Exit(1)
	}

	// Get the user-specified timezone or default to the local timezone
	tz := getTimeZone(*timezone)

	mdFiles, err := getMarkdownFiles(jekyllDir)
	if err != nil {
		logger.Println("Error getting markdown files")
		os.Exit(1)
	}

	for _, file := range mdFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			logger.Println("Error reading file: " + file)
			os.Exit(1)
		}

		frontMatter := extractFrontMatter(string(content))
		if frontMatter == nil {
			logger.Println("No front matter found in: " + file)
			continue
		}

		tomlData, err := convertFrontMatterToTOML(frontMatter, tz)
		if err != nil {
			// should display the error message with the file name and err.Error()
			logger.Println("Error: " + file + " - " + err.Error())
		}

		fmt.Println(string(tomlData))

		_ = tomlData

		// Write the TOML front matter to the Zola directory
		// zolaFile := filepath.Join(args["zolaDir"], "content", filepath.Base(file))
		// err = os.WriteFile(zolaFile, tomlData, 0644)

		// if err != nil {
		// 	logger.Println("Error writing file: " + zolaFile)
		// }
	}
}

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
func convertFrontMatterToTOML(frontMatter *string, loc *time.Location) ([]byte, error) {
	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(*frontMatter), &data)
	if err != nil {
		return nil, err
	}

	// Parse "date" field if it exists and is valid
	if dateStr, ok := data["date"].(string); ok {
		var t time.Time
		var err error

		for _, format := range dateFormats {
			if t, err = time.ParseInLocation(format, dateStr, loc); err == nil {
				data["date"] = t
				break
			}

			if t, err = time.Parse(format, dateStr); err == nil {
				data["date"] = t
				break
			}
		}

		if err != nil {
			return nil, fmt.Errorf("invalid date format in front matter: %s", dateStr)
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

// Get the timezone, defaulting to local timezone if not provided
func getTimeZone(tzName string) *time.Location {
	if tzName != "" {
		tz, err := time.LoadLocation(tzName)
		if err != nil {
			logger.Println("Invalid timezone specified, using local timezone instead.")
			return time.Local
		}
		return tz
	}
	return time.Local
}
