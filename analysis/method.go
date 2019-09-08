package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/sourcegraph/go-lsp"
)

type Method struct {
	location lsp.Location
}

func NewMethod(document *Document, parent *Symbol, node *phrase.Phrase) *Method {
	return &Method{
		location: document.GetNodeLocation(node),
	}
}
