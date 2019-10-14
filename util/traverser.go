package util

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
)

type Traverser struct {
	spine []*phrase.Phrase
	node  phrase.AstNode
	index int
}

func NewTraverser(node *phrase.Phrase) Traverser {
	return Traverser{
		node:  node,
		spine: []*phrase.Phrase{node},
		index: 0,
	}
}

func (s *Traverser) Advance() phrase.AstNode {
	node := s.Peek()
	if node != nil {
		s.index++
	}

	return node
}

func (s *Traverser) Peek() phrase.AstNode {
	p, ok := s.node.(*phrase.Phrase)
	if !ok {
		return nil
	}
	if s.index >= len(p.Children) {
		return nil
	}

	return p.Children[s.index]
}

func (s *Traverser) Descend() {
	p, ok := s.node.(*phrase.Phrase)
	if !ok {
		return
	}

	if s.index >= len(p.Children) {
		return
	}

	node := p.Children[s.index]
	if p, ok := node.(*phrase.Phrase); ok {
		s.spine = append(s.spine, p)
	}
	s.node = node
	s.index = 0
}

func (s *Traverser) SkipToken(tokenType lexer.TokenType) {
	next := s.Peek()
	for nextToken, ok := next.(*lexer.Token); ok && nextToken.Type == tokenType; {
		s.Advance()
		next = s.Peek()
		nextToken, ok = next.(*lexer.Token)
	}
}

func (s *Traverser) Clone() Traverser {
	return Traverser{
		node:  s.node,
		spine: s.spine,
		index: s.index,
	}
}
