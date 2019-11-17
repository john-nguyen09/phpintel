package analysis

import (
	"encoding/json"
	"sort"
	"sync"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Document contains information of documents
type Document struct {
	uri            string
	text           []rune
	lineOffsets    []int
	loadMu         sync.Mutex
	isLoaded       bool
	isOpen         bool
	variableTables []variableTable
	Children       []Symbol `json:"children"`
	classStack     []Symbol
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

func NewDocument(uri string, text string) *Document {
	document := &Document{
		uri:      uri,
		Children: []Symbol{},
	}
	document.SetText(text)

	return document
}

// Open sets a flag to indicate the document is open
func (s *Document) Open() {
	s.isOpen = true
}

// Close unsets the flag
func (s *Document) Close() {
	s.isOpen = false
}

// Load makes sure that symbols are available
func (s *Document) Load() {
	if !s.isLoaded {
		s.pushVariableTable()
		rootNode := parser.Parse(string(s.GetText()))
		scanForChildren(s, rootNode)
		s.isLoaded = true
	}
}

// LockToDo locks the document to do a thing
// the operation in thing should be synchronous, i.e.
// no goroutine to be spawned
func (s *Document) LockToDo(thing func(*Document)) {
	s.loadMu.Lock()
	defer s.loadMu.Unlock()
	thing(s)
}

// Release releases symbols to save memory
func (s *Document) Release() {
	s.variableTables = []variableTable{}
	s.Children = []Symbol{}
	s.classStack = []Symbol{}
	s.isLoaded = false
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
	s.text = []rune(text)
	s.lineOffsets = calculateLineOffsets(s.text, 0)
}

func calculateLineOffsets(text []rune, offset int) []int {
	n := 0
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
			lineOffsets = append(lineOffsets, n+offset)
		} else if c == '\n' {
			lineOffsets = append(lineOffsets, n+offset+1)
		}
	}
	if isLineStart {
		lineOffsets = append(lineOffsets, n)
	}
	return lineOffsets
}

func (s *Document) lineAt(offset int) int {
	return sort.Search(len(s.lineOffsets), func(i int) bool {
		return s.lineOffsets[i] > offset
	}) - 1
}

func (s *Document) offsetAtLine(line int) int {
	if line <= 0 || len(s.lineOffsets) < 1 {
		return 0
	}
	if line > len(s.lineOffsets)-1 {
		return s.lineOffsets[len(s.lineOffsets)-1]
	}
	return s.lineOffsets[line]
}

func (s *Document) positionAt(offset int) protocol.Position {
	line := s.lineAt(offset)
	return protocol.Position{
		Line:      line,
		Character: offset - s.lineOffsets[line],
	}
}

func (s *Document) offsetAtPosition(pos protocol.Position) int {
	offset := s.offsetAtLine(pos.Line) + pos.Character
	min := 0
	if offset < len(s.text) {
		min = offset
	} else {
		min = len(s.text)
	}
	if 0 > min {
		return 0
	} else {
		return min
	}
}

func (s *Document) NodeRange(node phrase.AstNode) protocol.Range {
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

	return protocol.Range{Start: s.positionAt(start), End: s.positionAt(end)}
}

// GetText is a getter for text
func (s *Document) GetText() []rune {
	return s.text
}

// GetNodeLocation retrieves the location of a phrase node
func (s *Document) GetNodeLocation(node phrase.AstNode) protocol.Location {
	return protocol.Location{
		URI:   protocol.DocumentURI(s.GetURI()),
		Range: s.NodeRange(node),
	}
}

func (s *Document) GetNodeText(node phrase.AstNode) string {
	switch node := node.(type) {
	case *lexer.Token:
		return s.GetTokenText(node)
	case *phrase.Phrase:
		return s.GetPhraseText(node)
	}

	return ""
}

func (s *Document) GetPhraseText(phrase *phrase.Phrase) string {
	firstToken, lastToken := util.FirstToken(phrase), util.LastToken(phrase)

	return string(s.text[firstToken.Offset : lastToken.Offset+lastToken.Length])
}

func (s *Document) GetTokenText(token *lexer.Token) string {
	return string(s.text[token.Offset : token.Offset+token.Length])
}

func (s *Document) addSymbol(other Symbol) {
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

// Even though the name indicates class but actually this will also
// return interface and trait
func (s *Document) getLastClass() Symbol {
	return s.classStack[len(s.classStack)-1]
}

func (s *Document) addClass(other Symbol) {
	switch instance := other.(type) {
	case *Class:
		s.classStack = append(s.classStack, instance)
	case *Interface:
		s.classStack = append(s.classStack, instance)
	case *Trait:
		s.classStack = append(s.classStack, instance)
	}
}

func (s *Document) SymbolAt(offset int) HasTypes {
	pos := s.positionAt(offset)
	return s.SymbolAtPos(pos)
}

func (s *Document) SymbolAtPos(pos protocol.Position) HasTypes {
	index := sort.Search(len(s.Children), func(i int) bool {
		location := s.Children[i].GetLocation()
		return util.IsInRange(pos, location.Range) <= 0
	})
	for _, symbol := range s.Children[index:] {
		inRange := util.IsInRange(pos, symbol.GetLocation().Range)
		if inRange > 0 {
			break
		}
		if hasTypes, ok := symbol.(HasTypes); ok && inRange == 0 {
			return hasTypes
		}
	}
	return nil
}

func (s *Document) ApplyChanges(changes []protocol.TextDocumentContentChangeEvent) {
	for _, change := range changes {
		start := change.Range.Start
		end := change.Range.End
		text := []rune(change.Text)

		startOffset := s.offsetAtPosition(start)
		endOffset := s.offsetAtPosition(end)
		s.text = append(s.text, s.text[0:startOffset]...)
		s.text = append(s.text, text...)
		s.text = append(s.text, s.text[endOffset:]...)

		newLineOffsets := s.lineOffsets[0:change.Range.Start.Line]
		lengthDiff := len(text) - (endOffset - startOffset)
		newLineOffsets = append(newLineOffsets, calculateLineOffsets(text, startOffset)[1:]...)
		endLineOffsets := s.lineOffsets[end.Line+1:]
		for _, endLineOffset := range endLineOffsets {
			newLineOffsets = append(newLineOffsets, endLineOffset+lengthDiff)
		}
		s.lineOffsets = newLineOffsets
	}
	s.isLoaded = false
}
