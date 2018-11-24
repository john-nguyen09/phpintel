package util

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/sourcegraph/go-lsp"
)

func GetPhraseText(phrase *phrase.Phrase, docText []rune) []rune {
	firstToken, lastToken := FirstToken(phrase), LastToken(phrase)

	return docText[firstToken.Offset : lastToken.Offset+lastToken.Length]
}

func GetTokenText(token *lexer.Token, docText []rune) []rune {
	return docText[token.Offset : token.Offset+token.Length]
}

func FirstToken(node phrase.AstNode) *lexer.Token {
	if t, ok := node.(*lexer.Token); ok {
		return t
	}

	if p, ok := node.(*phrase.Phrase); ok {
		for _, child := range p.Children {
			t := FirstToken(child)

			if t != nil {
				return t
			}
		}
	}

	return nil
}

func LastToken(node phrase.AstNode) *lexer.Token {
	if t, ok := node.(*lexer.Token); ok {
		return t
	}

	if p, ok := node.(*phrase.Phrase); ok {
		for i := len(p.Children) - 1; i >= 0; i-- {
			t := LastToken(p.Children[i])

			if t != nil {
				return t
			}
		}
	}

	return nil
}

func NodeRange(node phrase.AstNode, text []rune) lsp.Range {
	var start, end int

	switch node := node.(type) {
	case *lexer.Token:
		start = node.Offset
		end = node.Offset + node.Length
	case *phrase.Phrase:
		firstToken, lastToken := FirstToken(node), LastToken(node)

		start = firstToken.Offset
		end = lastToken.Offset + lastToken.Length
	}

	return lsp.Range{Start: ToPosition(start, text), End: ToPosition(end, text)}
}
