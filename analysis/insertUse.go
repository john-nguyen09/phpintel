package analysis

import (
	"index/suffixarray"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

var numLinesAfterUse = 2

type InsertUseContext struct {
	document     *Document
	firstInline  *sitter.Node
	namespaceDef *sitter.Node
	lastUse      *sitter.Node
}

func GetInsertUseContext(document *Document) InsertUseContext {
	insertUseCtx := InsertUseContext{
		document:     document,
		firstInline:  nil,
		namespaceDef: nil,
		lastUse:      nil,
	}
	traverser := util.NewTraverser(document.GetRootNode())
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "php_tag":
			if insertUseCtx.firstInline == nil {
				insertUseCtx.firstInline = child
			}
		case "namespace_definition":
			insertUseCtx.namespaceDef = child
		case "namespace_use_declaration":
			insertUseCtx.lastUse = child
		}
		child = traverser.Advance()
	}
	return insertUseCtx
}

func (i InsertUseContext) GetInsertAfterNode() *sitter.Node {
	if i.lastUse != nil {
		return i.lastUse
	}
	if i.namespaceDef != nil {
		return i.namespaceDef
	}
	if i.firstInline != nil {
		return i.firstInline
	}

	return nil
}

func (i InsertUseContext) GetInsertPosition() (protocol.Position, bool) {
	afterNode := i.GetInsertAfterNode()
	if afterNode != nil {
		return util.PointToPosition(afterNode.EndPoint()), true
	}
	return protocol.Position{}, false
}

func (i InsertUseContext) GetUseEdit(typeString TypeString, symbol Symbol, alias string) *protocol.TextEdit {
	if typeString.GetFQN() == "" {
		return nil
	}
	if insertedPosition, ok := i.GetInsertPosition(); ok {
		eol := i.document.detectedEOL
		afterNode := i.GetInsertAfterNode()
		text := eol

		if afterNode.Type() == "namespace_definition" {
			text += eol
		}

		text += getIndentation(i.document, afterNode) + "use "
		switch symbol.(type) {
		case *Function:
			text += "function "
		case *Const, *Define:
			text += "const "
		}
		text += typeString.GetFQN()[1:]
		if alias != "" {
			text += " as " + alias
		}
		text += ";" + getNewLine(i.document, afterNode)

		return &protocol.TextEdit{
			Range: protocol.Range{
				Start: insertedPosition,
				End:   insertedPosition,
			},
			NewText: text,
		}
	}
	return nil
}

func getIndentation(document *Document, node *sitter.Node) string {
	nodeStart := util.PointToPosition(node.StartPoint())
	startOffset := document.OffsetAtPosition(protocol.Position{
		Line:      nodeStart.Line,
		Character: 0,
	})
	return string(document.GetText()[startOffset:node.StartByte()])
}

func getNewLine(document *Document, node *sitter.Node) string {
	next := node.NextSibling()
	if next == nil {
		return document.detectedEOL
	}
	nodeEnd := node.EndByte()
	startNext := next.StartByte()
	index := suffixarray.New(document.GetText()[nodeEnd:startNext])
	numNewLines := len(index.Lookup([]byte(document.detectedEOL), -1))
	if numNewLines < numLinesAfterUse {
		newLines := ""
		for i := 0; i < (numLinesAfterUse - numNewLines); i++ {
			newLines += document.detectedEOL
		}
		return newLines
	}
	return ""
}
