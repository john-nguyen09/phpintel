package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	lsp "github.com/sourcegraph/go-lsp"
)

type Document struct {
	Uri      string `json:"uri"`
	text     []rune
	Children []Symbol `json:"children"`
}

func NewDocument(uri string, text []rune, rootNode *phrase.Phrase) *Document {
	document := &Document{
		Uri:      uri,
		text:     text,
		Children: []Symbol{},
	}

	ScanForChildren(document, rootNode)

	return document
}

func (s *Document) GetDocument() *Document {
	return s
}

func (s *Document) GetUri() string {
	return s.Uri
}

func (s *Document) GetText() []rune {
	return s.text
}

func (s *Document) GetNodeLocation(node *phrase.Phrase) lsp.Location {
	return lsp.Location{
		URI:   lsp.DocumentURI(s.GetUri()),
		Range: util.NodeRange(node, s.GetText()),
	}
}

func (s *Document) Consume(other Symbol) {
	s.Children = append(s.Children, other)
}
