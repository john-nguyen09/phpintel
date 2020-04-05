package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type PropertyAccess struct {
	Expression

	hasResolved bool
}

func newPropertyAccess(document *Document, node *ast.Node) (HasTypes, bool) {
	propertyAccess := &PropertyAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	expression := scanForExpression(document, firstChild)
	if expression != nil {
		propertyAccess.Scope = expression
	}

	propertyAccess.Name, propertyAccess.Location = readMemberName(document, traverser)
	return propertyAccess, true
}

func (s *PropertyAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *PropertyAccess) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	store := ctx.store
	for _, scopeType := range s.ResolveAndGetScope(ctx).Resolve() {
		for _, class := range store.GetClasses(scopeType.GetFQN()) {
			for _, property := range GetClassProperties(store, class, "$"+s.Name, NewSearchOptions()) {
				s.Type.merge(property.Types)
			}
		}
	}
	s.hasResolved = true
}

func (s *PropertyAccess) GetTypes() TypeComposite {
	return s.Type
}
