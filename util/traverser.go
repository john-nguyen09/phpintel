package util

import (
	"errors"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
)

type Traverser struct {
	node   phrase.AstNode
	index  int
	parent *Traverser
}

func NewTraverser(node *phrase.Phrase) *Traverser {
	return &Traverser{
		node:   node,
		index:  0,
		parent: nil,
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

func (s *Traverser) Descend() (*Traverser, error) {
	p, ok := s.node.(*phrase.Phrase)
	if !ok {
		return nil, errors.New("Cannot descend into token")
	}

	if s.index >= len(p.Children) {
		return nil, errors.New("This is outside of children")
	}

	return &Traverser{
		node:   p.Children[s.index],
		index:  0,
		parent: s,
	}, nil
}

func (s *Traverser) Ascend() (*Traverser, error) {
	if s.parent == nil {
		return nil, errors.New("Cannot ascend because has not been descended")
	}
	return s.parent, nil
}

func (s *Traverser) SkipToken(tokenType lexer.TokenType) {
	next := s.Peek()
	for nextToken, ok := next.(*lexer.Token); ok && nextToken.Type == tokenType; {
		s.Advance()
		next = s.Peek()
		nextToken, ok = next.(*lexer.Token)
	}
}

func (s *Traverser) Clone() *Traverser {
	return &Traverser{
		node:  s.node,
		index: s.index,
	}
}

func (s *Traverser) Traverse(visit func(phrase.AstNode) bool) {
	s.realTraverse(s.node, visit)
}

func (s *Traverser) realTraverse(node phrase.AstNode, visit func(phrase.AstNode) bool) {
	shouldAscend := visit(node)
	if !shouldAscend {
		return
	}
	if p, ok := node.(*phrase.Phrase); ok {
		for _, child := range p.Children {
			s.realTraverse(child, visit)
		}
	}
}
