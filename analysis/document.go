package analysis

import (
	"crypto/md5"
	"encoding/json"
	"regexp"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

var /* const */ wordRegex = regexp.MustCompile(`[a-zA-Z_\x80-\xff][\\a-zA-Z0-9_\x80-\xff]*$`)

// Document contains information of documents
type Document struct {
	uri         string
	rootNode    *phrase.Phrase
	text        []rune
	lineOffsets []int
	loadMu      sync.Mutex
	isOpen      bool
	detectedEOL string

	variableTables     []*VariableTable
	variableTableLevel int
	Children           []Symbol `json:"children"`
	hasTypesSymbols    []HasTypes
	argLists           []*ArgumentList
	classStack         []Symbol
	lastPhpDoc         *phpDocComment
	hasChanges         bool
	importTable        ImportTable
}

// VariableTable holds the range and the variables inside
type VariableTable struct {
	locationRange  protocol.Range
	variables      map[string]*Variable
	globalDeclares map[string]bool
	level          int
}

func newVariableTable(locationRange protocol.Range, level int) *VariableTable {
	return &VariableTable{
		locationRange:  locationRange,
		variables:      map[string]*Variable{},
		globalDeclares: map[string]bool{},
		level:          level,
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

func (vt *VariableTable) canReferenceGlobal(name string) bool {
	if _, ok := vt.globalDeclares[name]; ok {
		return true
	}
	return false
}

func (vt VariableTable) setReferenceGlobal(name string) {
	vt.globalDeclares[name] = true
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
		uri:                uri,
		Children:           []Symbol{},
		variableTableLevel: 0,
		hasChanges:         true,
		importTable:        newImportTable(),
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

func (s *Document) IsOpen() bool {
	return s.isOpen
}

func (s *Document) ResetState() {
	s.Children = []Symbol{}
	s.hasTypesSymbols = []HasTypes{}
	s.argLists = []*ArgumentList{}
	s.variableTableLevel = 0
	s.variableTables = []*VariableTable{}
	s.classStack = []Symbol{}
	s.lastPhpDoc = nil
	s.rootNode = nil
}

func (s *Document) GetRootNode() *phrase.Phrase {
	if s.rootNode == nil {
		s.rootNode = parser.Parse(string(s.GetText()))
	}
	return s.rootNode
}

// Load makes sure that symbols are available
func (s *Document) Load() {
	if !s.hasChanges {
		return
	}
	s.ResetState()
	s.hasChanges = false
	rootNode := s.GetRootNode()
	s.pushVariableTable(rootNode)
	scanForChildren(s, rootNode)
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
	s.lineOffsets, s.detectedEOL = calculateLineOffsets(s.text, 0)
}

func calculateLineOffsets(text []rune, offset int) ([]int, string) {
	n := 0
	length := len(text)
	isLineStart := true
	lineOffsets := []int{}
	var c rune
	eol := "\n"
	stopDetectingEol := false

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
				if !stopDetectingEol {
					eol = "\r\n"
					stopDetectingEol = true
				}
			} else {
				if !stopDetectingEol {
					eol = "\r"
					stopDetectingEol = true
				}
			}
			lineOffsets = append(lineOffsets, n+offset)
		} else if c == '\n' {
			lineOffsets = append(lineOffsets, n+offset+1)
		}
	}
	if isLineStart {
		lineOffsets = append(lineOffsets, n)
	}
	return lineOffsets, eol
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

func (s *Document) OffsetAtPosition(pos protocol.Position) int {
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

func (s *Document) errorRange(err *phrase.ParseError) protocol.Range {
	if len(err.Children) == 0 {
		return protocol.Range{
			Start: s.positionAt(err.Unexpected.Offset),
			End:   s.positionAt(err.Unexpected.Offset + err.Unexpected.Length),
		}
	}

	firstToken := err.Children[0].(*lexer.Token)
	lastToken := err.Children[len(err.Children)-1].(*lexer.Token)
	return protocol.Range{
		Start: s.positionAt(firstToken.Offset),
		End:   s.positionAt(lastToken.Offset + lastToken.Length),
	}
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
	if firstToken == nil || lastToken == nil {
		return ""
	}

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
	if other == nil {
		debug.PrintStack()
	}
	if argList, ok := other.(*ArgumentList); ok {
		s.argLists = append(s.argLists, argList)
		return
	}
	if h, ok := other.(HasTypes); ok {
		s.hasTypesSymbols = append(s.hasTypesSymbols, h)
		return
	}
	s.Children = append(s.Children, other)
}

func (s *Document) pushVariableTable(node *phrase.Phrase) {
	s.variableTables = append(s.variableTables, newVariableTable(s.nodeRange(node), s.variableTableLevel))
	s.variableTableLevel++
}

func (s *Document) getCurrentVariableTable() *VariableTable {
	return s.variableTables[len(s.variableTables)-1]
}

// GetVariableTableAt returns the closest variable table which is in range
func (s *Document) GetVariableTableAt(pos protocol.Position) *VariableTable {
	// The algorithm is going reverse until found one, since parent tables
	// are appear first therefore going backward makes sure that parent is
	// the last to be considered
	for i := range s.variableTables {
		varTable := s.variableTables[len(s.variableTables)-1-i]
		if util.IsInRange(pos, varTable.locationRange) == 0 {
			return varTable
		}
	}
	// This might not be position because the whole file is a variable table
	// so only if the pos is outside of the file
	return nil
}

func (s *Document) pushVariable(variable *Variable) {
	variableTable := s.getCurrentVariableTable()
	currentVariable := variableTable.get(variable.Name)
	if currentVariable != nil {
		variable.mergeTypesWithVariable(currentVariable)
	}
	if variableTable.level == 0 || variableTable.canReferenceGlobal(variable.Name) {
		variable.canReferenceGlobal = true
	}
	variableTable.add(variable)
}

// Even though the name indicates class but actually this will also
// return interface and trait
func (s *Document) getLastClass() Symbol {
	if len(s.classStack) == 0 {
		return nil
	}
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

func (s *Document) NodeSpineAt(offset int) util.NodeStack {
	found := util.NodeStack{}
	traverser := util.NewTraverser(s.GetRootNode())
	traverser.Traverse(func(node phrase.AstNode, spine []*phrase.Phrase) bool {
		if t, ok := node.(*lexer.Token); ok && offset > t.Offset && offset <= (t.Offset+t.Length) {
			found = append(spine[:0:0], spine...)
			return false
		}
		return true
	})
	return found
}

// HasTypesAt is an interface to SymbolAtPos but with offset
func (s *Document) HasTypesAt(offset int) HasTypes {
	pos := s.positionAt(offset)
	return s.HasTypesAtPos(pos)
}

// HasTypesAtPos returns a HasTypes symbol at the position
func (s *Document) HasTypesAtPos(pos protocol.Position) HasTypes {
	index := sort.Search(len(s.hasTypesSymbols), func(i int) bool {
		location := s.hasTypesSymbols[i].GetLocation()
		return util.IsInRange(pos, location.Range) <= 0
	})
	var previousHasTypes HasTypes = nil
	for _, symbol := range s.hasTypesSymbols[index:] {
		inRange := util.IsInRange(pos, symbol.GetLocation().Range)
		if inRange < 0 {
			break
		}
		if hasTypes, ok := symbol.(HasTypes); ok && inRange == 0 {
			previousHasTypes = hasTypes
		}
	}
	return previousHasTypes
}

func (s *Document) Lock() {
	s.loadMu.Lock()
}

func (s *Document) Unlock() {
	s.loadMu.Unlock()
}

// HasTypesBeforePos returns a HasTypes before the position
func (s *Document) HasTypesBeforePos(pos protocol.Position) HasTypes {
	return s.hasTypesBeforePos(pos)
}

func (s *Document) hasTypesBeforePos(pos protocol.Position) HasTypes {
	index := sort.Search(len(s.hasTypesSymbols), func(i int) bool {
		location := s.hasTypesSymbols[i].GetLocation()
		return util.IsInRange(pos, location.Range) <= 0
	})
	for i := index; i >= 0; i-- {
		h := s.hasTypesSymbols[i]
		inRange := util.IsInRange(pos, h.GetLocation().Range)
		if hasTypes, ok := h.(HasTypes); ok && (inRange > 0 || (inRange == 0 && pos == h.GetLocation().Range.End)) {
			return hasTypes
		}
	}
	return nil
}

func (s *Document) WordAtPos(pos protocol.Position) string {
	offset := s.OffsetAtPosition(pos)
	lineNumber := sort.SearchInts(s.lineOffsets, offset) - 1
	if lineNumber < 0 {
		lineNumber = 0
	}
	if lineNumber < len(s.lineOffsets)-1 && offset == s.lineOffsets[lineNumber+1] {
		lineNumber++
	}
	lineSubString := string(s.GetText()[s.lineOffsets[lineNumber]:offset])
	return wordRegex.FindString(lineSubString)
}

// ArgumentListAndFunctionCallAt returns an ArgumentList and FunctionCall at the position
func (s *Document) ArgumentListAndFunctionCallAt(pos protocol.Position) (*ArgumentList, HasParamsResolvable) {
	// log.Printf("ArgumentListAndFunctionCallAt: %p", s)
	index := sort.Search(len(s.argLists), func(i int) bool {
		location := s.argLists[i].GetLocation()
		return util.IsInRange(pos, location.Range) <= 0
	})
	var hasParamsResolvable HasParamsResolvable = nil
	var argumentList *ArgumentList = nil
	for _, argList := range s.argLists[index:] {
		isInRange := util.IsInRange(pos, argList.GetLocation().Range)
		if isInRange == 0 {
			argumentList = argList
		}
		if isInRange < 0 {
			break
		}
	}
	if argumentList != nil {
		hasTypes := s.hasTypesBeforePos(argumentList.GetLocation().Range.Start)
		if resolvable, ok := hasTypes.(HasParamsResolvable); ok {
			hasParamsResolvable = resolvable
		}
	}
	return argumentList, hasParamsResolvable
}

// ApplyChanges applies the changes to line offsets and text
func (s *Document) ApplyChanges(changes []protocol.TextDocumentContentChangeEvent) {
	defer util.TimeTrack(time.Now(), "ApplyChanges")
	// log.Printf("ApplyChanges: %p", s)
	s.hasChanges = true
	for _, change := range changes {
		start := change.Range.Start
		end := change.Range.End
		text := []rune(change.Text)

		startOffset := s.OffsetAtPosition(start)
		endOffset := s.OffsetAtPosition(end)
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
		offsets, eol := calculateLineOffsets(text, startOffset)
		s.detectedEOL = eol
		newLineOffsets = append(newLineOffsets, offsets[1:]...)
		if end.Line+1 < len(s.lineOffsets) {
			endLineOffsets := s.lineOffsets[end.Line+1:]
			for _, endLineOffset := range endLineOffsets {
				newLineOffsets = append(newLineOffsets, endLineOffset+lengthDiff)
			}
		}
		s.lineOffsets = newLineOffsets
	}
	s.Load()
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
	if endOfPhpDoc.Line < start.Line && endOfPhpDoc.Line == (start.Line-1) {
		return s.lastPhpDoc
	}
	return nil
}

func (s *Document) getGlobalVariable(name string) *GlobalVariable {
	for _, child := range s.Children {
		if globalVariable, ok := child.(*GlobalVariable); ok && globalVariable.GetName() == name {
			return globalVariable
		}
	}
	return nil
}

func (s *Document) GetMD5Hash() []byte {
	hasher := md5.New()
	hasher.Write(util.RunesToUTF8(s.GetText()))
	return hasher.Sum(nil)
}

func (s *Document) GetImportTable() ImportTable {
	return s.importTable
}

func (s *Document) setNamespace(namespace *Namespace) {
	s.importTable.setNamespace(namespace)
}
