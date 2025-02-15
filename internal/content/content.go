package content

import (
	"fmt"
	"regexp"
)

// Combines TOML front matter with markdown content
func CombineFrontMatterAndContent(tomlData []byte, content []byte) string {
	ContentProcessing(&content)

	re := regexp.MustCompile(`(?s)---\n(.*?)\n---`)
	content = re.ReplaceAll(content, []byte(""))

	return fmt.Sprintf("+++\n%s+++%s", tomlData, content)
}

func ContentProcessing(content *[]byte) {
	// Correct the <!--more--> tag
	re := regexp.MustCompile(`(?i)<!--\s*more\s*-->`)
	if re.Match(*content) {
		*content = re.ReplaceAll(*content, []byte("\n<!--more-->"))
	}

	// TODO: logic to process relative urls/paths in markdown content
}
