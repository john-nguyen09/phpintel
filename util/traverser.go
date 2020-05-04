package util

import (
	"errors"
	"strings"

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
		node:   s.node,
		index:  s.index,
		parent: s.parent,
	}
}

type VisitorContext struct {
	ShouldAscend bool
	AscendNode   phrase.AstNode
}

type Visitor func(phrase.AstNode, []*phrase.Phrase) VisitorContext

func (s *Traverser) Traverse(visit Visitor) {
	spine := []*phrase.Phrase{}
	s.realTraverse(s.node, spine, visit)
}

func (s *Traverser) realTraverse(node phrase.AstNode, spine []*phrase.Phrase, visit Visitor) {
	ctx := visit(node, spine)
	if !ctx.ShouldAscend {
		return
	}
	if ctx.AscendNode != nil {
		node = ctx.AscendNode
	}
	if p, ok := node.(*phrase.Phrase); ok {
		spine = append(spine, p)
		for _, child := range p.Children {
			s.realTraverse(child, spine, visit)
		}
		spine = spine[:len(spine)-1]
	}
}

type NodeStack []*phrase.Phrase

func (s *NodeStack) Parent() phrase.Phrase {
	if len(*s) == 0 {
		return phrase.Phrase{
			Type: phrase.Unknown,
			Children: []phrase.AstNode{
				lexer.Token{
					Type: lexer.Undefined,
				},
			},
		}
	}
	var p *phrase.Phrase
	p, *s = (*s)[len(*s)-1], (*s)[:len(*s)-1]
	return *p
}

func (s NodeStack) String() string {
	strs := []string{}
	for p := s.Parent(); p.Type != phrase.Unknown; p = s.Parent() {
		strs = append(strs, p.Type.String())
	}
	return strings.Join(strs, ", ")
}
