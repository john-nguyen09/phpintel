package analysis

import (
	"crypto/sha1"
	"encoding/json"
	"regexp"
	"runtime/debug"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
)

var /* const */ wordRegex = regexp.MustCompile(`[a-zA-Z_\x80-\xff][\\a-zA-Z0-9_\x80-\xff]*$`)

// Document contains information of documents
type Document struct {
	uri         string
	tree        *sitter.Tree
	text        []byte
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
	children       []*VariableTable
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

func (vt *VariableTable) addChild(child *VariableTable) {
	vt.children = append(vt.children, child)
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

func NewDocument(uri string, text []byte) *Document {
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
	s.importTable = newImportTable()
}

func (s *Document) GetRootNode() *sitter.Node {
	if s.tree == nil {
		p := sitter.NewParser()
		p.SetLanguage(php.GetLanguage())
		s.tree = p.Parse(nil, s.GetText())
	}
	return s.tree.RootNode()
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
func (s *Document) SetText(text []byte) {
	s.text = text
	s.lineOffsets, s.detectedEOL = calculateLineOffsets(s.text, 0)
}

func calculateLineOffsets(text []byte, offset int) ([]int, string) {
	n := 0
	isLineStart := true
	lineOffsets := []int{}
	var r rune
	var size int
	eol := "\n"
	stopDetectingEol := false

	for len(text) > 0 {
		r, size = utf8.DecodeRune(text)
		if isLineStart {
			lineOffsets = append(lineOffsets, n)
			isLineStart = false
		}
		if r == '\r' {
			if len(text) > 0 {
				nextR, nextSize := utf8.DecodeRune(text[size:])
				if nextR == '\n' {
					text = text[size+nextSize:]
					n += size + nextSize
					if !stopDetectingEol {
						eol = "\r\n"
						stopDetectingEol = true
					}
					lineOffsets = append(lineOffsets, n+offset)
					continue
				}
			} else {
				if !stopDetectingEol {
					eol = "\r"
					stopDetectingEol = true
				}
			}
			lineOffsets = append(lineOffsets, n+offset)
		} else if r == '\n' {
			lineOffsets = append(lineOffsets, n+offset+1)
		}
		text = text[size:]
		n += size
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

func (s *Document) nodeRange(node *sitter.Node) protocol.Range {
	return protocol.Range{Start: util.PointToPosition(node.StartPoint()), End: util.PointToPosition(node.EndPoint())}
}

func (s *Document) errorRange(err *sitter.Node) protocol.Range {
	return s.nodeRange(err)
}

// GetText is a getter for text
func (s *Document) GetText() []byte {
	return s.text
}

// GetNodeLocation retrieves the location of a phrase node
func (s *Document) GetNodeLocation(node *sitter.Node) protocol.Location {
	return protocol.Location{
		URI:   protocol.DocumentURI(s.GetURI()),
		Range: s.nodeRange(node),
	}
}

func (s *Document) GetNodeText(node *sitter.Node) string {
	return node.Content(s.GetText())
}

func (s *Document) GetPhraseText(phrase *sitter.Node) string {
	return s.GetNodeText(phrase)
}

func (s *Document) GetTokenText(token *sitter.Node) string {
	return s.GetNodeText(token)
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
		if len(s.argLists) > 0 {
			i := len(s.argLists) - 1
			lastArgList := s.argLists[i]
			if util.IsInRange(argList.GetLocation().Range.Start, lastArgList.GetLocation().Range) == 0 {
				s.argLists = append(s.argLists[:i], append([]*ArgumentList{argList}, s.argLists[i:]...)...)
				return
			}
		}
		s.argLists = append(s.argLists, argList)
		return
	}
	if h, ok := other.(HasTypes); ok {
		s.hasTypesSymbols = append(s.hasTypesSymbols, h)
		return
	}
	s.Children = append(s.Children, other)
}

func (s *Document) pushVariableTable(node *sitter.Node) {
	newVarTable := newVariableTable(s.nodeRange(node), s.variableTableLevel)
	if s.variableTableLevel > 0 {
		s.getCurrentVariableTable().addChild(newVarTable)
	}
	s.variableTables = append(s.variableTables, newVarTable)
	s.variableTableLevel++
}

func (s *Document) popVariableTable() *VariableTable {
	length := len(s.variableTables)
	last, poppedVariableTables := s.variableTables[length-1], s.variableTables[:length-1]
	s.variableTables = poppedVariableTables
	return last
}

func (s *Document) getCurrentVariableTable() *VariableTable {
	return s.variableTables[len(s.variableTables)-1]
}

// GetVariableTableAt returns the closest variable table which is in range
func (s *Document) GetVariableTableAt(pos protocol.Position) *VariableTable {
	// The first element is supposed to always be there because it represents
	// the scope of the whole document
	lastFoundVarTable := s.variableTables[0]
	for _, varTable := range lastFoundVarTable.children {
		if util.IsInRange(pos, varTable.locationRange) == 0 {
			lastFoundVarTable = varTable
		}
	}
	return lastFoundVarTable
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

func (s *Document) getClassScopeAtSymbol(symbol Symbol) string {
	class := s.getClassAtPos(symbol.GetLocation().Range.Start)
	if class == nil {
		return ""
	}

	switch v := class.(type) {
	case *Class:
		return v.Name.GetFQN()
	case *Interface:
		return v.Name.GetFQN()
	}
	return ""
}

func (s *Document) getClassAtPos(pos protocol.Position) Symbol {
	index := sort.Search(len(s.classStack), func(i int) bool {
		return util.IsInRange(pos, s.classStack[i].GetLocation().Range) <= 0
	})
	if index >= len(s.classStack) {
		return nil
	}
	return s.classStack[index]
}

func (s *Document) NodeSpineAt(offset int) util.NodeStack {
	found := util.NodeStack{}
	traverser := util.NewTraverser(s.GetRootNode())
	traverser.Traverse(func(node *sitter.Node, spine []*sitter.Node) bool {
		if node.ChildCount() == 0 && offset > int(node.StartByte()) && offset <= int(node.EndByte()) {
			found = append(spine[:0:0], spine...)
			found = append(found, node)
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
	if index >= len(s.hasTypesSymbols) {
		index = len(s.hasTypesSymbols) - 1
	}
	for i := index; i >= 0; i-- {
		h := s.hasTypesSymbols[i]
		inRange := util.IsInRange(pos, h.GetLocation().Range)
		hRange := h.GetLocation().Range
		if inRange > 0 || (inRange == 0 && pos == hRange.End && hRange.Start != hRange.End) {
			return h
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
	if index < len(s.argLists) && util.IsInRange(pos, s.argLists[index].GetLocation().Range) == 0 {
		argumentList = s.argLists[index]
	} else {
		for _, arg := range s.argLists[index:] {
			isInRange := util.IsInRange(pos, arg.GetLocation().Range)
			if isInRange > 0 {
				break
			}
			if isInRange == 0 {
				argumentList = arg
				break
			}
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
	// log.Printf("ApplyChanges: %p", s)
	start := time.Now()
	s.hasChanges = true
	for _, change := range changes {
		start := change.Range.Start
		end := change.Range.End
		text := []byte(change.Text)

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
	util.TimeTrack(start, "contentChanges")
	start = time.Now()
	for _, change := range changes {
		start := change.Range.Start
		end := change.Range.End
		text := []byte(change.Text)
		startOffset := s.OffsetAtPosition(start)
		endOffset := s.OffsetAtPosition(end)
		rangeLength := endOffset - startOffset

		s.tree.Edit(sitter.EditInput{
			StartIndex:  uint32(startOffset),
			OldEndIndex: uint32(startOffset) + uint32(rangeLength),
			NewEndIndex: uint32(startOffset) + uint32(len(text)),
			StartPoint:  util.PositionToPoint(start),
			OldEndPoint: util.PositionToPoint(s.positionAt(startOffset + rangeLength)),
			NewEndPoint: util.PositionToPoint(s.positionAt(startOffset + len(text))),
		})
	}
	p := sitter.NewParser()
	p.SetLanguage(php.GetLanguage())
	oldTree := s.tree
	s.tree = p.Parse(oldTree, s.GetText())
	util.TimeTrack(start, "editASTTree")
	start = time.Now()
	s.Load()
	util.TimeTrack(start, "Load")
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

func (s *Document) GetHash() []byte {
	hash := sha1.Sum(s.GetText())
	return hash[:]
}

func (s *Document) GetImportTable() ImportTable {
	return s.importTable
}

func (s *Document) setNamespace(namespace *Namespace) {
	s.importTable.setNamespace(namespace)
}
