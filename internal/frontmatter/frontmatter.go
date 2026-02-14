package frontmatter

import (
	"bytes"
	"fmt"
)

var (
	openDelim  = []byte("---\n")
	closeDelim = []byte("\n---")
)

// findBounds returns the start and end byte offsets of the front matter
// content (excluding delimiters) within raw. Returns an error if the
// opening or closing delimiter is not found.
func findBounds(raw []byte) (start, end int, err error) {
	s := bytes.Index(raw, openDelim)
	if s == -1 {
		return 0, 0, fmt.Errorf("no opening front matter delimiter")
	}
	s += len(openDelim)

	e := bytes.Index(raw[s:], closeDelim)
	if e == -1 {
		return 0, 0, fmt.Errorf("no closing front matter delimiter")
	}
	return s, s + e, nil
}

// Extract returns the front matter content between the first pair of
// "---\n" / "\n---" delimiters.
func Extract(content []byte) ([]byte, error) {
	start, end, err := findBounds(content)
	if err != nil {
		return nil, err
	}
	return content[start:end], nil
}

// Strip removes the first front matter block (including delimiters)
// from content and returns the remainder.
func Strip(content []byte) []byte {
	s := bytes.Index(content, openDelim)
	if s == -1 {
		return content
	}
	bodyStart := s + len(openDelim)
	e := bytes.Index(content[bodyStart:], closeDelim)
	if e == -1 {
		return content
	}
	// skip past "\n---" (4 bytes)
	after := content[bodyStart+e+len(closeDelim):]
	return append(content[:s], after...)
}
