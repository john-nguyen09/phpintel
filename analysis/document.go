package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	lsp "github.com/sourcegraph/go-lsp"
)

type Document struct {
	Symbol
	uri      string
	text     []rune
	children []Symbol
}

func NewDocument(uri string, text []rune, rootNode *phrase.Phrase) *Document {
	document := &Document{
		uri:      uri,
		text:     text,
		children: []Symbol{},
	}

	ScanForChildren(document, rootNode)

	return document
}

func (s *Document) GetDocument() *Document {
	return s
}

func (s *Document) GetChildren() []Symbol {
	return s.children
}

func (s *Document) GetUri() string {
	return s.uri
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
	s.children = append(s.children, other)
}
