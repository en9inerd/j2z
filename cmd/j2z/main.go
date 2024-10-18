package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

func main() {
	var mdFiles []string
	args := getAndCheckArgs()

	mdFiles = getMarkdownFiles(args["jekyllDir"])

	for _, file := range mdFiles {
		fmt.Println(file)

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

		// Write the TOML front matter to the Zola directory
		zolaFile := filepath.Join(args["zolaDir"], filepath.Base(file))
		err = os.WriteFile(zolaFile, tomlData, 0644)

		if err != nil {
			fmt.Println("Error writing file: " + zolaFile)
		}
	}
}

func getAndCheckArgs() map[string]string {
	var jekyllDir string
	var zolaDir string

	var args = os.Args[1:]
	if len(args) != 2 {
		fmt.Println(
			"No arguments provided" + "\n\n" + "Usage: j2z [jekyll dir] [zola dir]",
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

func getMarkdownFiles(jekyllDir string) []string {
	var files []string

	err := filepath.Walk(jekyllDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error walking through the jekyll directory")
	}

	return files
}

func extractFrontMatter(content string) *string {
	re := regexp.MustCompile(`(?s)---\n(.*?)\n---`)
	match := re.FindStringSubmatch(content)

	if len(match) == 0 {
		return nil
	}

	return &match[1]
}

func convertFrontMatterToTOML(frontMatter string) ([]byte, error) {
	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(frontMatter), &data)

	if err != nil {
		return nil, err
	}

	var tomlData []byte
	tomlData, err = toml.Marshal(data)

	if err != nil {
		return nil, err
	}

	return tomlData, nil
}
