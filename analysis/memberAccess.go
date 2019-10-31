package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
)

func readMemberName(document *Document, node *phrase.Phrase) string {
	return document.GetNodeText(node)
}
