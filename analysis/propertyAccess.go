package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type PropertyAccess struct {
	Expression

	hasResolved bool
}

var _ HasTypesHasScope = (*PropertyAccess)(nil)

func newPropertyAccess(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	propertyAccess := &PropertyAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		expression := scanForExpression(a, document, p)
		if expression != nil {
			propertyAccess.Scope = expression
		}
	}

	propertyAccess.Name, propertyAccess.Location = readMemberName(a, document, traverser)
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

func (s *PropertyAccess) GetScopeTypes() TypeComposite {
	if s.Scope != nil {
		return s.Scope.GetTypes()
	}
	return newTypeComposite()
}

func (s *PropertyAccess) MemberName() string {
	name := []rune(s.Name)
	if len(name) > 0 && name[0] != '$' {
		name = append([]rune{'$'}, name...)
	}
	return string(name)
}
