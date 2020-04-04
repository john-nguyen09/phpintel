package ast

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// Node is a wrapper of sitter.Node to reduce its overhead
type Node struct {
	n        *sitter.Node
	t        string
	children []*Node
}

// FromSitterNode creates a node from the sitter.Node
func FromSitterNode(n *sitter.Node) *Node {
	return &Node{
		n:        n,
		children: make([]*Node, n.ChildCount()),
	}
}

// ChildCount calls sitter.Node.ChildCount()
func (n *Node) ChildCount() uint32 {
	return n.n.ChildCount()
}

// Type returns the type of the node
func (n *Node) Type() string {
	if n.t == "" {
		n.t = n.n.Type()
	}
	return n.t
}

// Child returns a node and cache it
func (n *Node) Child(idx int) *Node {
	if n.children[idx] == nil {
		n.children[idx] = FromSitterNode(n.n.Child(idx))
	}
	return n.children[idx]
}

// Cursor returns the sitter.TreeCursor for the node
func (n *Node) Cursor() *sitter.TreeCursor {
	return sitter.NewTreeCursor(n.n)
}

// StartPoint calls sitter.Node.StartPoint
func (n *Node) StartPoint() sitter.Point {
	return n.n.StartPoint()
}

// EndPoint calls sitter.Node.StartPoint
func (n *Node) EndPoint() sitter.Point {
	return n.n.EndPoint()
}

// StartByte calls sitter.Node.StartByte
func (n *Node) StartByte() uint32 {
	return n.n.StartByte()
}

// EndByte calls sitter.Node.EndByte
func (n *Node) EndByte() uint32 {
	return n.n.EndByte()
}

// Content calls sitter.Node.Content
func (n *Node) Content(input []byte) string {
	return n.n.Content(input)
}

// ChildByFieldName calls sitter.Node.ChildByFieldName
func (n *Node) ChildByFieldName(name string) *Node {
	child := n.n.ChildByFieldName(name)
	if child == nil {
		return nil
	}
	return FromSitterNode(child)
}

// NextSibling calls sitter.Node.NextSibling
func (n *Node) NextSibling() *Node {
	sib := n.n.NextSibling()
	if sib == nil {
		return nil
	}
	return FromSitterNode(sib)
}

// IsMissing calls sitter.Node.IsMissing
func (n *Node) IsMissing() bool {
	return n.n.IsMissing()
}

// PrevSibling calls sitter.Node.PrevSibling
func (n *Node) PrevSibling() *Node {
	sib := n.n.PrevSibling()
	if sib == nil {
		return nil
	}
	return FromSitterNode(sib)
}

// Parent calls sitter.Node.Parent
func (n *Node) Parent() *Node {
	par := n.n.Parent()
	if par == nil {
		return nil
	}
	return FromSitterNode(par)
}
