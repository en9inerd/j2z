package content

import (
	"strings"
	"testing"
)

func TestNormalizeMoreTag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "canonical tag untouched",
			input: "before\n<!--more-->\nafter",
			want:  "before\n<!--more-->\nafter",
		},
		{
			name:  "extra whitespace normalized",
			input: "before\n<!-- more -->\nafter",
			want:  "before\n<!--more-->\nafter",
		},
		{
			name:  "case insensitive",
			input: "before\n<!-- MORE -->\nafter",
			want:  "before\n<!--more-->\nafter",
		},
		{
			name:  "mixed case with spaces",
			input: "before\n<!--  More  -->\nafter",
			want:  "before\n<!--more-->\nafter",
		},
		{
			name:  "inline tag gets newline prepended",
			input: "end of paragraph.<!--more-->\nnext paragraph",
			want:  "end of paragraph.\n<!--more-->\nnext paragraph",
		},
		{
			name:  "already on own line no double newline",
			input: "paragraph\n<!--more-->\nnext",
			want:  "paragraph\n<!--more-->\nnext",
		},
		{
			name:  "no more tag",
			input: "just regular <!-- comment --> text",
			want:  "just regular <!-- comment --> text",
		},
		{
			name:  "multiple more tags",
			input: "a<!-- more -->b<!-- MORE -->c",
			want:  "a\n<!--more-->b\n<!--more-->c",
		},
		{
			name:  "no html comments at all",
			input: "plain text without any tags",
			want:  "plain text without any tags",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMoreTag([]byte(tt.input))
			if string(got) != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestConvertLiquidHighlight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple highlight block",
			input: "before\n{% highlight ruby %}\nputs 'hello'\n{% endhighlight %}\nafter",
			want:  "before\n```ruby\nputs 'hello'\n```\nafter",
		},
		{
			name:  "highlight with no lang",
			input: "{% highlight %}\ncode\n{% endhighlight %}",
			want:  "```\ncode\n```",
		},
		{
			name:  "no highlight tags",
			input: "just plain text",
			want:  "just plain text",
		},
		{
			name:  "highlight with other liquid tags nearby",
			input: "{% include header.html %}\n{% highlight go %}\nfmt.Println()\n{% endhighlight %}",
			want:  "{% include header.html %}\n```go\nfmt.Println()\n```",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertLiquidHighlight([]byte(tt.input))
			if string(got) != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestCombineFrontMatterAndContent(t *testing.T) {
	toml := []byte("title = \"Test\"\n")
	content := []byte("---\ntitle: Test\n---\n\nBody text here.")

	result := CombineFrontMatterAndContent(toml, content)

	if !strings.HasPrefix(result, "+++\n") {
		t.Error("result should start with TOML delimiter +++")
	}
	if !strings.Contains(result, "title = \"Test\"") {
		t.Error("result should contain TOML front matter")
	}
	if !strings.Contains(result, "Body text here.") {
		t.Error("result should contain body text")
	}
	if strings.Contains(result, "---") {
		t.Error("result should not contain YAML delimiters")
	}
}
