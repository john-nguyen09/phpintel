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
	variableTables []VariableTable
	Children       []Symbol `json:"children"`
	classStack     []Symbol
	lastPhpDoc     *phpDocComment
}

// VariableTable holds the range and the variables inside
type VariableTable struct {
	locationRange protocol.Range
	variables     map[string]*Variable
}

func newVariableTable(locationRange protocol.Range) VariableTable {
	return VariableTable{
		locationRange: locationRange,
		variables:     map[string]*Variable{},
	}
}

func (vt *VariableTable) add(variable *Variable) {
	vt.variables[variable.Name] = variable
}

func (vt *VariableTable) get(name string) *Variable {
	if variable, ok := vt.variables[name]; ok {
		return variable
	}
	return nil
}

// GetVariables returns all the variables in the table
func (vt *VariableTable) GetVariables() map[string]*Variable {
	return vt.variables
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
		rootNode := parser.Parse(string(s.GetText()))
		s.pushVariableTable(rootNode)
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
	s.variableTables = []VariableTable{}
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
	}
	return min
}

func (s *Document) nodeRange(node phrase.AstNode) protocol.Range {
	var start, end int

	switch node := node.(type) {
	case *lexer.Token:
		start = node.Offset
		end = node.Offset + node.Length
	case *phrase.Phrase:
		firstToken, lastToken := util.FirstToken(node), util.LastToken(node)
		if firstToken != nil || lastToken != nil {
			start = firstToken.Offset
			end = lastToken.Offset + lastToken.Length
		}
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
		Range: s.nodeRange(node),
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
	s.lastPhpDoc = nil
	if doc, ok := other.(*phpDocComment); ok {
		s.lastPhpDoc = doc
		return
	}
	s.Children = append(s.Children, other)
}

func (s *Document) pushVariableTable(node *phrase.Phrase) {
	s.variableTables = append(s.variableTables, newVariableTable(s.nodeRange(node)))
}

func (s *Document) getCurrentVariableTable() VariableTable {
	return s.variableTables[len(s.variableTables)-1]
}

// GetVariableTableAt returns the closest variable table which is in range
func (s *Document) GetVariableTableAt(pos protocol.Position) VariableTable {
	found := s.variableTables[0] // First one is always the document
	foundOne := false
	// The algorithm is that the first one is in range means the next ones
	// might also be in range so it goes on until no more in-range ones
	// and return the last one in range
	for _, varTable := range s.variableTables {
		if util.IsInRange(pos, varTable.locationRange) == 0 {
			found = varTable
			foundOne = true
		} else {
			if foundOne {
				return found
			}
		}
	}
	// If this is reached then pos is either outside the range of document variableTable
	// therefore returning the document variableTable
	return found
}

func (s *Document) pushVariable(variable *Variable) {
	variableTable := s.getCurrentVariableTable()
	currentVariable := variableTable.get(variable.Name)
	if currentVariable != nil {
		variable.mergeTypesWithVariable(currentVariable)
	}
	variableTable.add(variable)
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

// SymbolAt is an interface to SymbolAtPos but with offset
func (s *Document) SymbolAt(offset int) HasTypes {
	pos := s.positionAt(offset)
	return s.SymbolAtPos(pos)
}

// SymbolAtPos returns a HasTypes symbol at the position
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

// ApplyChanges applies the changes to line offsets and text
func (s *Document) ApplyChanges(changes []protocol.TextDocumentContentChangeEvent) {
	for _, change := range changes {
		start := change.Range.Start
		end := change.Range.End
		text := []rune(change.Text)

		startOffset := s.offsetAtPosition(start)
		endOffset := s.offsetAtPosition(end)
		newText := append(s.text[:0:0], s.text[0:startOffset]...)
		newText = append(newText, text...)
		newText = append(newText, s.text[endOffset:]...)
		s.text = newText

		min := start.Line + 1
		if min > len(s.lineOffsets) {
			min = len(s.lineOffsets)
		}
		newLineOffsets := append(s.lineOffsets[:0:0], s.lineOffsets[0:min]...)
		lengthDiff := len(text) - (endOffset - startOffset)
		newLineOffsets = append(newLineOffsets, calculateLineOffsets(text, startOffset)[1:]...)
		if end.Line+1 < len(s.lineOffsets) {
			endLineOffsets := s.lineOffsets[end.Line+1:]
			for _, endLineOffset := range endLineOffsets {
				newLineOffsets = append(newLineOffsets, endLineOffset+lengthDiff)
			}
		}
		s.lineOffsets = newLineOffsets
	}
}

func (s *Document) getLines() []string {
	lines := []string{}
	text := s.GetText()
	lineOffsets := s.lineOffsets

	start, lineOffsets := s.lineOffsets[0], lineOffsets[1:]
	for index, lineOffset := range lineOffsets {
		line := ""
		if index != len(lineOffsets)-1 {
			line = string(text[start:lineOffset])
		} else {
			line = string(text[start : lineOffset-1])
		}
		lines = append(lines, line)
		start = lineOffset
	}
	if start == len(text) {
		lines = append(lines, "")
	} else {
		lines = append(lines, string(text[start:len(text)]))
	}
	return lines
}

func (s *Document) getValidPhpDoc(location protocol.Location) *phpDocComment {
	if s.lastPhpDoc == nil {
		return nil
	}
	endOfPhpDoc := s.lastPhpDoc.GetLocation().Range.End
	start := location.Range.Start
	if endOfPhpDoc.Line < start.Line && endOfPhpDoc.Line >= (start.Line-2) {
		return s.lastPhpDoc
	}
	return nil
}
