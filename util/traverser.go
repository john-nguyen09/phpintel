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

// NodeStack contains the token and its parents
type NodeStack struct {
	parents []*phrase.Phrase
	token   *lexer.Token
}

// Push adds the node to the stack
func (s *NodeStack) Push(node *phrase.Phrase) {
	s.parents = append(s.parents, node)
}

// Pop pops the node from the stack or nil if stack is empty
func (s *NodeStack) Pop() *phrase.Phrase {
	if len(s.parents) == 0 {
		return nil
	}
	var p *phrase.Phrase
	p, s.parents = s.parents[len(s.parents)-1], s.parents[:len(s.parents)-1]
	return p
}

// SetParents sets the parents of the stack
func (s *NodeStack) SetParents(parents []*phrase.Phrase) *NodeStack {
	s.parents = parents
	return s
}

// SetToken sets the token to the stack
func (s *NodeStack) SetToken(token *lexer.Token) *NodeStack {
	s.token = token
	return s
}

// Parent returns the parent and mutate the stack
func (s *NodeStack) Parent() phrase.Phrase {
	p := s.Phrase()
	s.Pop()
	return p
}

// Phrase returns the phrase of the stack
func (s *NodeStack) Phrase() phrase.Phrase {
	if len(s.parents) == 0 {
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
	p = s.parents[len(s.parents)-1]
	return *p
}

// Token returns the token of the stack
func (s NodeStack) Token() lexer.Token {
	if s.token == nil {
		return lexer.Token{Type: lexer.Undefined}
	}
	return *s.token
}

// String returns the string representation of a NodeStack
func (s NodeStack) String() string {
	strs := []string{}

	if s.token != nil {
		strs = append(strs, s.token.String())
	}

	for p := s.Parent(); p.Type != phrase.Unknown; p = s.Parent() {
		strs = append(strs, p.Type.String())
	}
	return strings.Join(strs, ", ")
}
