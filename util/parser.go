package util

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	sitter "github.com/smacker/go-tree-sitter"
)

func FirstToken(node *sitter.Node) *sitter.Node {
	if node.ChildCount() == 0 {
		return node
	}
	return node.Child(0)
}

func LastToken(node *sitter.Node) *sitter.Node {
	if node.ChildCount() == 0 {
		return node
	}

	return node.Child(int(node.ChildCount()) - 1)
}

func IsOfPhraseType(node phrase.AstNode, phraseType phrase.PhraseType) (*phrase.Phrase, bool) {
	p, ok := node.(*phrase.Phrase)
	if !ok {
		return nil, false
	}
	return p, p.Type == phraseType
}

func IsOfPhraseTypes(node phrase.AstNode, phraseTypes []phrase.PhraseType) (*phrase.Phrase, bool) {
	p, ok := node.(*phrase.Phrase)
	if !ok {
		return nil, false
	}
	for _, phraseType := range phraseTypes {
		if p.Type == phraseType {
			return p, true
		}
	}
	return nil, false
}

func PointToPosition(p sitter.Point) protocol.Position {
	return protocol.Position{
		Line:      int(p.Row),
		Character: int(p.Column),
	}
}

func PositionToPoint(p protocol.Position) sitter.Point {
	return sitter.Point{
		Row:    uint32(p.Line),
		Column: uint32(p.Character),
	}
}
