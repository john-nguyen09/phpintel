package analysis

import (
	"encoding/json"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	lsp "github.com/sourcegraph/go-lsp"
)

// Document contains information of documents
type Document struct {
	uri            string
	text           string
	variableTables []variableTable
	Children       []Symbol `json:"children"`
}

type variableTable map[string]*Variable

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

func newDocument(uri string, text string, rootNode *phrase.Phrase) *Document {
	document := &Document{
		uri:      uri,
		text:     text,
		Children: []Symbol{},
	}
	document.pushVariableTable()

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
func (s *Document) GetText() string {
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

func (s *Document) pushVariableTable() {
	s.variableTables = append(s.variableTables, variableTable{})
}

func (s *Document) getCurrentVariableTable() variableTable {
	return s.variableTables[len(s.variableTables)-1]
}

func (s *Document) pushVariable(variable *Variable) {
	variableTable := s.getCurrentVariableTable()
	if currentVariable, ok := variableTable[variable.Name]; ok {
		variable.mergeTypesWithVariable(currentVariable)
	}
	variableTable[variable.Name] = variable
}
