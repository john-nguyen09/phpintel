package util

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
)

// FirstToken returns the first terminal node (leave)
func FirstToken(node phrase.AstNode) *lexer.Token {
	if t, ok := node.(*lexer.Token); ok {
		return t
	}
	if e, ok := node.(*phrase.ParseError); ok {
		return e.Unexpected
	}
	if p, ok := node.(*phrase.Phrase); ok && len(p.Children) != 0 {
		return FirstToken(p.Children[0])
	}
	return nil
}

// LastToken returns the last terminal node (leave)
func LastToken(node phrase.AstNode) *lexer.Token {
	if t, ok := node.(*lexer.Token); ok {
		return t
	}
	if e, ok := node.(*phrase.ParseError); ok {
		return e.Unexpected
	}
	if p, ok := node.(*phrase.Phrase); ok && len(p.Children) != 0 {
		return LastToken(p.Children[len(p.Children)-1])
	}
	return nil
}

// IsOfPhraseType checks if a node is the given type
func IsOfPhraseType(node phrase.AstNode, phraseType phrase.PhraseType) (*phrase.Phrase, bool) {
	p, ok := node.(*phrase.Phrase)
	if !ok {
		return nil, false
	}
	return p, p.Type == phraseType
}

// IsOfPhraseTypes checks if a node is one of the given types
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
