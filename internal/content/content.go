package content

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/en9inerd/j2z/internal/frontmatter"
)

// CombineFrontMatterAndContent combines TOML front matter with markdown content.
func CombineFrontMatterAndContent(tomlData []byte, content []byte) string {
	processContent(&content)
	content = frontmatter.Strip(content)
	return fmt.Sprintf("+++\n%s+++%s", tomlData, content)
}

func processContent(content *[]byte) {
	*content = normalizeMoreTag(*content)
	*content = convertLiquidHighlight(*content)
	*content = warnLiquidIncludes(*content)
}

// convertLiquidHighlight converts Jekyll's {% highlight lang %} ... {% endhighlight %}
// blocks into standard fenced code blocks (```lang ... ```).
func convertLiquidHighlight(content []byte) []byte {
	openTag := []byte("{%")
	var result []byte

	for len(content) > 0 {
		idx := bytes.Index(content, openTag)
		if idx == -1 {
			result = append(result, content...)
			break
		}

		closeIdx := bytes.Index(content[idx:], []byte("%}"))
		if closeIdx == -1 {
			result = append(result, content...)
			break
		}

		inner := bytes.TrimSpace(content[idx+2 : idx+closeIdx])
		if lang, ok := bytes.CutPrefix(inner, []byte("highlight")); ok {
			lang = bytes.TrimSpace(lang)
			endTag := []byte("{% endhighlight %}")
			endIdx := bytes.Index(content[idx+closeIdx+2:], endTag)
			if endIdx == -1 {
				result = append(result, content[:idx+closeIdx+2]...)
				content = content[idx+closeIdx+2:]
				continue
			}

			codeBlock := content[idx+closeIdx+2 : idx+closeIdx+2+endIdx]
			codeBlock = bytes.TrimPrefix(codeBlock, []byte("\n"))
			codeBlock = bytes.TrimSuffix(codeBlock, []byte("\n"))

			result = append(result, content[:idx]...)
			result = append(result, []byte("```")...)
			result = append(result, lang...)
			result = append(result, '\n')
			result = append(result, codeBlock...)
			result = append(result, '\n')
			result = append(result, []byte("```")...)

			content = content[idx+closeIdx+2+endIdx+len(endTag):]
		} else {
			result = append(result, content[:idx+closeIdx+2]...)
			content = content[idx+closeIdx+2:]
		}
	}

	return result
}

// warnLiquidIncludes logs a warning for any {% include ... %} tags found
// in the content, as these have no Zola equivalent.
func warnLiquidIncludes(content []byte) []byte {
	search := content
	for {
		idx := bytes.Index(search, []byte("{%"))
		if idx == -1 {
			break
		}
		closeIdx := bytes.Index(search[idx:], []byte("%}"))
		if closeIdx == -1 {
			break
		}
		inner := bytes.TrimSpace(search[idx+2 : idx+closeIdx])
		if bytes.HasPrefix(inner, []byte("include")) {
			slog.Warn("unsupported Liquid tag found (no Zola equivalent)",
				"tag", string(bytes.TrimSpace(inner)))
		}
		search = search[idx+closeIdx+2:]
	}
	return content
}

// normalizeMoreTag replaces any variant of the <!--more--> tag
// (case-insensitive, with optional whitespace) with the canonical form.
// It ensures the tag sits on its own line, adding a newline before it
// only when the preceding character is not already a newline.
func normalizeMoreTag(content []byte) []byte {
	lower := bytes.ToLower(content)
	var result []byte

	for len(content) > 0 {
		idx := bytes.Index(lower, []byte("<!--"))
		if idx == -1 {
			result = append(result, content...)
			break
		}

		endIdx := bytes.Index(lower[idx:], []byte("-->"))
		if endIdx == -1 {
			result = append(result, content...)
			break
		}

		inner := bytes.TrimSpace(lower[idx+4 : idx+endIdx])
		if string(inner) == "more" {
			result = append(result, content[:idx]...)
			if len(result) > 0 && result[len(result)-1] != '\n' {
				result = append(result, '\n')
			}
			result = append(result, []byte("<!--more-->")...)
		} else {
			result = append(result, content[:idx+endIdx+3]...)
		}
		content = content[idx+endIdx+3:]
		lower = lower[idx+endIdx+3:]
	}

	return result
}
