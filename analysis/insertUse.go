package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type InsertUseContext struct {
	document     *Document
	firstInline  *phrase.Phrase
	namespaceDef *phrase.Phrase
	lastUse      *phrase.Phrase
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
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.InlineText:
				if insertUseCtx.firstInline == nil {
					insertUseCtx.firstInline = p
				}
			case phrase.NamespaceDefinition:
				insertUseCtx.namespaceDef = p
			case phrase.NamespaceUseDeclaration:
				insertUseCtx.lastUse = p
			}
		}
		child = traverser.Advance()
	}
	return insertUseCtx
}

func (i InsertUseContext) GetInsertAfterNode() *phrase.Phrase {
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
		return i.document.positionAt(lastToken.Offset + lastToken.Length), true
	}
	return protocol.Position{}, false
}

func (i InsertUseContext) GetUseEdit(typeString TypeString, symbol Symbol, alias string) *protocol.TextEdit {
	if insertedPosition, ok := i.GetInsertPosition(); ok {
		eol := i.document.detectedEOL
		afterNode := i.GetInsertAfterNode()
		text := eol

		if afterNode.Type == phrase.NamespaceDefinition {
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
		text += ";"

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

func getIndentation(document *Document, node *phrase.Phrase) string {
	firstToken := util.FirstToken(node)
	tokenStartPosition := document.positionAt(firstToken.Offset)
	startOffset := document.offsetAtPosition(protocol.Position{
		Line:      tokenStartPosition.Line,
		Character: 0,
	})
	return string(document.GetText()[startOffset:firstToken.Offset])
}