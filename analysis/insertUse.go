package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

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
		lastToken := util.LastToken(afterNode)
		return util.PointToPosition(lastToken.EndPoint()), true
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
		text += ";\n"

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
	return "\t"
}
