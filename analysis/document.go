package analysis

import (
	"encoding/json"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	lsp "github.com/sourcegraph/go-lsp"
)

// Document contains information of documents
type Document struct {
	uri      string
	text     []rune
	Children []Symbol `json:"children"`
}

// MarshalJSON is used for json.Marshal
func (s *Document) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		URI      string
		Children []Symbol
	}{
		URI:      s.uri,
		Children: s.Children,
	})
}

func newDocument(uri string, text []rune, rootNode *phrase.Phrase) *Document {
	document := &Document{
		uri:      uri,
		text:     text,
		Children: []Symbol{},
	}

	scanForChildren(document, rootNode)

	return document
}

func (s *Document) getDocument() *Document {
	return s
}

// GetURI is a getter for uri
func (s *Document) GetURI() string {
	return s.uri
}

// GetText is a getter for text
func (s *Document) GetText() []rune {
	return s.text
}

// GetNodeLocation retrieves the location of a phrase node
func (s *Document) GetNodeLocation(node phrase.AstNode) lsp.Location {
	return lsp.Location{
		URI:   lsp.DocumentURI(s.GetURI()),
		Range: util.NodeRange(node, s.GetText()),
	}
}

func (s *Document) consume(other Symbol) {
	s.Children = append(s.Children, other)
}
