package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
)

// ProcessExpressionStatement is a proxy to scan for other types
func ProcessExpressionStatement(document *Document, parent symbolBlock, node *phrase.Phrase) Symbol {
	scanForChildren(parent, node)

	return nil
}
