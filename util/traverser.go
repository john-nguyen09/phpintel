package util

import (
	"errors"

	sitter "github.com/smacker/go-tree-sitter"
)

type Traverser struct {
	node   *sitter.Node
	index  int
	parent *Traverser
}

func NewTraverser(node *sitter.Node) *Traverser {
	return &Traverser{
		node:   node,
		index:  0,
		parent: nil,
	}
}

func (s *Traverser) Advance() *sitter.Node {
	node := s.Peek()
	if node != nil {
		s.index++
	}

	return node
}

func (s *Traverser) Peek() *sitter.Node {
	if s.index >= int(s.node.ChildCount()) {
		return nil
	}

	return s.node.Child(s.index)
}

func (s *Traverser) Descend() (*Traverser, error) {
	if s.index >= int(s.node.ChildCount()) {
		return nil, errors.New("This is outside of children")
	}

	return &Traverser{
		node:   s.node.Child(s.index),
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

func (s *Traverser) SkipToken(tokenType string) {
	next := s.Peek()
	for next != nil && next.Type() == tokenType {
		s.Advance()
		next = s.Peek()
	}
}

func (s *Traverser) Clone() *Traverser {
	return &Traverser{
		node:  s.node,
		index: s.index,
	}
}

type Visitor func(*sitter.Node, []*sitter.Node) bool

func (s *Traverser) Traverse(visit Visitor) {
	spine := []*sitter.Node{}
	s.realTraverse(s.node, spine, visit)
}

func (s *Traverser) realTraverse(node *sitter.Node, spine []*sitter.Node, visit Visitor) {
	shouldAscend := visit(node, spine)
	if !shouldAscend {
		return
	}
	childCount := int(node.ChildCount())
	if childCount > 0 {
		spine = append(spine, node)
		for i := 0; i < childCount; i++ {
			s.realTraverse(node.Child(i), spine, visit)
		}
		spine = spine[:len(spine)-1]
	}
}

type NodeStack []*sitter.Node

func (s *NodeStack) Parent() *sitter.Node {
	if len(*s) == 0 {
		return nil
	}
	var p *sitter.Node
	p, *s = (*s)[len((*s))-1], (*s)[:len((*s))-1]
	return p
}
