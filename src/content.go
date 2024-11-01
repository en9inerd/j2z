package main

import (
	"fmt"
	"regexp"
)

// Combines TOML front matter with markdown content
func combineFrontMatterAndContent(tomlData []byte, content []byte) string {
	contentProcessing(&content)

	re := regexp.MustCompile(`(?s)---\n(.*?)\n---`)
	content = re.ReplaceAll(content, []byte(""))

	return fmt.Sprintf("+++\n%s+++%s", tomlData, content)
}

func contentProcessing(content *[]byte) {
}
