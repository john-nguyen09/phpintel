package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type Parenthesised struct {
	Expression
	hasResolved bool
}

var _ HasTypes = (*Parenthesised)(nil)

func newParenthesised(doc *Document, node *sitter.Node) (HasTypes, bool) {
	s := &Parenthesised{
		Expression: Expression{
			Location: doc.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		expr := scanForExpression(doc, child)
		if expr != nil {
			s.Scope = expr
		}
		child = traverser.Advance()
	}
	return s, true
}

func (s *Parenthesised) GetLocation() protocol.Location {
	return s.Location
}

func (s *Parenthesised) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	s.hasResolved = true
	if s.Scope != nil {
		s.Scope.Resolve(ctx)
		s.Type.merge(s.Scope.GetTypes())
	}
}

func (s *Parenthesised) GetTypes() TypeComposite {
	return s.Type
}
