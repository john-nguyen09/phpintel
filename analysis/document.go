package analysis

import (
	"crypto/sha1"
	"encoding/json"
	"regexp"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/john-nguyen09/phpintel/analysis/ast"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

var /* const */ wordRegex = regexp.MustCompile(`[\\a-zA-Z_\x80-\xff][\\a-zA-Z0-9_\x80-\xff]*$`)

// Document contains information of documents
type Document struct {
	uri         string
	injector    *ast.Injector
	text        []byte
	lineOffsets []int
	loadMu      sync.Mutex
	isOpen      bool
	detectedEOL string

	variableTables     []*VariableTable
	variableTableLevel int
	Children           []Symbol `json:"children"`
	classStack         []Symbol
	lastPhpDoc         *phpDocComment
	hasChanges         bool
	importTables       []*ImportTable
	insertUseContext   *InsertUseContext

	blockStack []BlockSymbol
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
		importTables:       []*ImportTable{},
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
	s.variableTableLevel = 0
	s.variableTables = []*VariableTable{}
	s.classStack = []Symbol{}
	s.lastPhpDoc = nil
	s.importTables = []*ImportTable{}
	s.injector = nil
}

func (s *Document) GetRootNode() *ast.Node {
	if s.injector == nil {
		s.injector = ast.NewPHPInjector(s.GetText())
	}
	return s.injector.MainRootNode()
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

	if len(s.importTables) == 0 {
		s.pushImportTable(rootNode)
	}
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

func (s *Document) pushImportTable(node *ast.Node) {
	s.importTables = append(s.importTables, newImportTable(s, node))
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

func (s *Document) nodeRange(node *ast.Node) protocol.Range {
	return protocol.Range{Start: util.PointToPosition(node.StartPoint()), End: util.PointToPosition(node.EndPoint())}
}

// GetText is a getter for text
func (s *Document) GetText() []byte {
	return s.text
}

// GetNodeLocation retrieves the location of a phrase node
func (s *Document) GetNodeLocation(node *ast.Node) protocol.Location {
	return protocol.Location{
		URI:   protocol.DocumentURI(s.GetURI()),
		Range: s.nodeRange(node),
	}
}

func (s *Document) GetNodeText(node *ast.Node) string {
	return node.Content(s.GetText())
}

func (s *Document) addSymbol(other Symbol) {
	s.lastPhpDoc = nil
	if doc, ok := other.(*phpDocComment); ok {
		s.lastPhpDoc = doc
		return
	}
	if s.currentBlock() != nil {
		s.currentBlock().addChild(other)
	} else {
		s.Children = append(s.Children, other)
	}
}

func (s *Document) pushBlock(block BlockSymbol) {
	s.blockStack = append(s.blockStack, block)
}

func (s *Document) popBlock() {
	s.blockStack = s.blockStack[:len(s.blockStack)-1]
}

func (s *Document) currentBlock() BlockSymbol {
	if len(s.blockStack) > 0 {
		return s.blockStack[len(s.blockStack)-1]
	}
	return nil
}

func (s *Document) pushVariableTable(node *ast.Node) {
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
	s.variableTableLevel--
	return last
}

func (s *Document) getCurrentVariableTable() *VariableTable {
	return s.variableTables[len(s.variableTables)-1]
}

// GetVariableTableAt returns the closest variable table which is in range
func (s *Document) GetVariableTableAt(pos protocol.Position) *VariableTable {
	var traverseAndFind func(*VariableTable) *VariableTable
	traverseAndFind = func(vt *VariableTable) *VariableTable {
		if len(vt.children) == 0 {
			return vt
		}
		for _, child := range vt.children {
			if util.IsInRange(pos, child.locationRange) == 0 {
				return traverseAndFind(child)
			}
		}
		return vt
	}
	// The first element is supposed to always be there because it represents
	// the scope of the whole document
	lastFoundVarTable := s.variableTables[0]
	return traverseAndFind(lastFoundVarTable)
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
	cursor := s.GetRootNode().Cursor()
	found := util.NodeStack{}
	uOffset := uint32(offset) - 1
	found = append(found, s.GetRootNode())
	for cursor.GoToFirstChildForByte(uOffset) != -1 {
		found = append(found, ast.FromSitterNode(cursor.CurrentNode()))
	}
	return found
}

// HasTypesAt is an interface to SymbolAtPos but with offset
func (s *Document) HasTypesAt(offset int) HasTypes {
	pos := s.positionAt(offset)
	return s.HasTypesAtPos(pos)
}

// HasTypesAtPos returns a HasTypes symbol at the position
func (s *Document) HasTypesAtPos(pos protocol.Position) HasTypes {
	var result HasTypes = nil
	t := newTraverser()
	t.traverseDocument(s, func(t *traverser, s Symbol) {
		relativePos := util.IsInRange(pos, s.GetLocation().Range)
		if relativePos == 0 {
			if h, ok := s.(HasTypes); ok {
				result = h
			}
		} else if relativePos < 0 {
			t.shouldStop = true
		} else if relativePos > 0 {
			t.stopDescent = true
		}
	})
	return result
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
	var result HasTypes = nil
	t := newTraverser()
	t.traverseDocument(s, func(t *traverser, s Symbol) {
		relativePos := util.IsInRange(pos, s.GetLocation().Range)
		if relativePos >= 0 {
			if h, ok := s.(HasTypes); ok {
				result = h
			}
		} else if relativePos < 0 {
			t.shouldStop = true
		}
	})
	return result
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
	var argumentList *ArgumentList = nil
	t := newTraverser()
	t.traverseDocument(s, func(t *traverser, s Symbol) {
		relativePos := util.IsInRange(pos, s.GetLocation().Range)
		if relativePos == 0 {
			if args, ok := s.(*ArgumentList); ok {
				argumentList = args
			}
		} else if relativePos < 0 {
			t.shouldStop = true
		} else if relativePos > 0 {
			t.stopDescent = true
		}
	})
	var hasParamsResolvable HasParamsResolvable = nil
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

		rangeLength := endOffset - startOffset
		oldEndIndex := startOffset + rangeLength
		newEndIndex := startOffset + len(text)
		edit := sitter.EditInput{
			StartIndex:  uint32(startOffset),
			OldEndIndex: uint32(oldEndIndex),
			NewEndIndex: uint32(newEndIndex),
			StartPoint:  util.PositionToPoint(start),
			OldEndPoint: util.PositionToPoint(s.positionAt(oldEndIndex)),
			NewEndPoint: util.PositionToPoint(s.positionAt(newEndIndex)),
		}
		if s.injector != nil {
			s.injector = s.injector.Edit(edit, s.GetText())
		}
	}
	util.TimeTrack(start, "contentChanges")
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

func (s *Document) currImportTable() *ImportTable {
	if len(s.importTables) == 0 {
		s.pushImportTable(s.GetRootNode())
	}
	return s.importTables[len(s.importTables)-1]
}

// ImportTableAtPos finds the importTable at the position
func (s *Document) ImportTableAtPos(pos protocol.Position) *ImportTable {
	index := sort.Search(len(s.importTables), func(i int) bool {
		return util.ComparePos(pos, s.importTables[i].start) <= 0
	})
	if index == 0 {
		return s.importTables[0]
	}
	return s.importTables[index-1]
}

func (s *Document) setNamespace(namespace *Namespace) {
	s.currImportTable().setNamespace(namespace)
}
