package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type Encapsulated struct {
	Expression
	hasResolved bool
}

func analyseEncapsulatedExpression(document *Document, node *sitter.Node) (HasTypes, bool) {
	en := &Encapsulated{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		ex := scanForExpression(document, child)
		if ex != nil {
			en.Scope = ex
		}
		child = traverser.Advance()
	}
	return en, true
}

func (s *Encapsulated) GetLocation() protocol.Location {
	return s.Location
}

func (s *Encapsulated) GetTypes() TypeComposite {
	return s.Type
}

func (s *Encapsulated) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	s.Type.merge(s.ResolveAndGetScope(ctx))
	s.hasResolved = true
}
