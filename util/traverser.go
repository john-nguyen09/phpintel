package util

import (
	"errors"
	"strings"

	"github.com/john-nguyen09/phpintel/analysis/ast"
)

type Traverser struct {
	node   *ast.Node
	index  int
	parent *Traverser
}

func NewTraverser(node *ast.Node) *Traverser {
	return &Traverser{
		node:   node,
		index:  0,
		parent: nil,
	}
}

func (s *Traverser) Advance() *ast.Node {
	node := s.Peek()
	if node != nil {
		s.index++
	}

	return node
}

func (s *Traverser) Peek() *ast.Node {
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

type VisitorContext struct {
	ShouldAscend bool
	AscendNode   *ast.Node
}

type Visitor func(*ast.Node, []*ast.Node) VisitorContext

func (s *Traverser) Traverse(visit Visitor) {
	spine := []*ast.Node{}
	s.realTraverse(s.node, spine, visit)
}

func (s *Traverser) realTraverse(node *ast.Node, spine []*ast.Node, visit Visitor) {
	ctx := visit(node, spine)
	if !ctx.ShouldAscend {
		return
	}
	if ctx.AscendNode != nil {
		node = ctx.AscendNode
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

type NodeStack []*ast.Node

func (s *NodeStack) Parent() *ast.Node {
	if len(*s) == 0 {
		return nil
	}
	var p *ast.Node
	p, *s = (*s)[len((*s))-1], (*s)[:len((*s))-1]
	return p
}

func (s NodeStack) String() string {
	strs := []string{}
	for p := s.Parent(); p != nil; p = s.Parent() {
		strs = append(strs, p.Type())
	}
	return strings.Join(strs, ", ")
}
