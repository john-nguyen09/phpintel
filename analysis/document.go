package analysis

import (
	"encoding/json"
	"sort"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	lsp "github.com/sourcegraph/go-lsp"
)

// Document contains information of documents
type Document struct {
	uri            string
	text           string
	lineOffsets    []int
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
		Children: []Symbol{},
	}
	document.SetText(text)
	document.pushVariableTable()

	// TODO: Remove rootNode dependency
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

// SetText is a setter for text, at the same time update line offsets
func (s *Document) SetText(text string) {
	s.text = text
	s.calculateLineOffsets()
}

func (s *Document) calculateLineOffsets() {
	n := 0
	text := []rune(s.GetText())
	length := len(text)
	isLineStart := true
	lineOffsets := []int{}
	var c rune

	for n := 0; n < length; n++ {
		c = text[n]
		if isLineStart {
			lineOffsets = append(lineOffsets, n)
			isLineStart = false
		}
		if c == '\r' {
			n++
			if n < length && text[n] == '\n' {
				n++
			}
			isLineStart = true
			continue
		} else if c == '\n' {
			isLineStart = true
		}
	}
	if isLineStart {
		lineOffsets = append(lineOffsets, n)
	}
	s.lineOffsets = lineOffsets
}

func (s *Document) lineAt(offset int) int {
	return sort.Search(len(s.lineOffsets), func(i int) bool {
		return s.lineOffsets[i] > offset
	}) - 1
}

func (s *Document) positionAt(offset int) lsp.Position {
	line := s.lineAt(offset)
	return lsp.Position{
		Line:      line,
		Character: offset - s.lineOffsets[line],
	}
}

func (s *Document) NodeRange(node phrase.AstNode) lsp.Range {
	var start, end int

	switch node := node.(type) {
	case *lexer.Token:
		start = node.Offset
		end = node.Offset + node.Length
	case *phrase.Phrase:
		firstToken, lastToken := util.FirstToken(node), util.LastToken(node)

		start = firstToken.Offset
		end = lastToken.Offset + lastToken.Length
	}

	return lsp.Range{Start: s.positionAt(start), End: s.positionAt(end)}
}

// GetText is a getter for text
func (s *Document) GetText() string {
	return s.text
}

// GetNodeLocation retrieves the location of a phrase node
func (s *Document) GetNodeLocation(node phrase.AstNode) lsp.Location {
	return lsp.Location{
		URI:   lsp.DocumentURI(s.GetURI()),
		Range: s.NodeRange(node),
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
