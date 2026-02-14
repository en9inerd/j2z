package frontmatter

import (
	"bytes"
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "simple front matter",
			input: "---\ntitle: Hello\n---\nbody",
			want:  "title: Hello",
		},
		{
			name:  "multi-line front matter",
			input: "---\ntitle: Hello\ndate: 2024-01-01\ntags: [a, b]\n---\nbody text",
			want:  "title: Hello\ndate: 2024-01-01\ntags: [a, b]",
		},
		{
			name:    "no opening delimiter",
			input:   "just some text",
			wantErr: true,
		},
		{
			name:    "no closing delimiter",
			input:   "---\ntitle: Hello\nno closing",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:  "empty front matter",
			input: "---\n\n---\nbody",
			want:  "",
		},
		{
			name:  "front matter with triple dashes in body",
			input: "---\ntitle: Test\n---\nbody with --- in it",
			want:  "title: Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Extract([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strip front matter",
			input: "---\ntitle: Hello\n---\nbody text",
			want:  "\nbody text",
		},
		{
			name:  "no front matter",
			input: "just body text",
			want:  "just body text",
		},
		{
			name:  "no closing delimiter",
			input: "---\ntitle: Hello\nno closing",
			want:  "---\ntitle: Hello\nno closing",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "only front matter",
			input: "---\ntitle: Hello\n---",
			want:  "",
		},
		{
			name:  "dashes embedded in content without newline prefix",
			input: "prefix---\ntitle: Hello\n---\nbody",
			want:  "prefix\nbody",
		},
		{
			name:  "front matter at start with newline body",
			input: "---\ntitle: Test\n---\n\nparagraph 1\n\nparagraph 2",
			want:  "\n\nparagraph 1\n\nparagraph 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Strip([]byte(tt.input))
			if !bytes.Equal(got, []byte(tt.want)) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
