package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/sourcegraph/go-lsp"
)

// Property contains information for properties
type Property struct {
	location lsp.Location
}

func newProperty(document *Document, parent *Symbol, node *phrase.Phrase) *Property {
	return &Property{
		location: document.GetNodeLocation(node),
	}
}
