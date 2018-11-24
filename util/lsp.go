package util

import (
	"strings"

	"github.com/sourcegraph/go-lsp"
)

func ToPosition(offset int, text []rune) lsp.Position {
	startAt := len(text)

	if offset < startAt {
		startAt = offset
	}

	lastNewLine := strings.LastIndex(string(text[0:startAt]), "\n")
	character := offset - (lastNewLine + 1)

	line := 0
	if offset > 0 {
		line = strings.Count(string(text[0:startAt]), "\n")
	}

	return lsp.Position{Line: line, Character: character}
}
