package util

import (
	"strings"

	"github.com/sourcegraph/go-lsp"
)

// substring without additional allocation but still safe
func substring(s string, start int, end int) string {
	startByteIndex := 0
	i := 0
	for j := range s {
		if i == start {
			startByteIndex = j
		}
		if i == end {
			return s[startByteIndex:j]
		}
		i++
	}
	return s[startByteIndex:]
}

func ToPosition(offset int, text string) lsp.Position {
	startAt := len(text)

	if offset < startAt {
		startAt = offset
		text = substring(text, 0, startAt)
	}
	lastNewLine := strings.LastIndex(text, "\n")
	character := offset - (lastNewLine + 1)

	line := 0
	if offset > 0 {
		line = strings.Count(text, "\n")
	}

	return lsp.Position{Line: line, Character: character}
}
