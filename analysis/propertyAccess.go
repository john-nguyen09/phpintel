package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type PropertyAccess struct {
	MemberAccessExpression

	hasResolved bool
}

var _ HasTypesHasScope = (*PropertyAccess)(nil)
var _ MemberAccess = (*PropertyAccess)(nil)

func newPropertyAccess(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	propertyAccess := &PropertyAccess{
		MemberAccessExpression: MemberAccessExpression{
			Expression: Expression{},
		},
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
	q := ctx.query
	currentClass := ctx.document.GetClassScopeAtSymbol(s.Scope)
	for _, scopeType := range s.ResolveAndGetScope(ctx).Resolve() {
		for _, class := range q.GetClasses(scopeType.GetFQN()) {
			for _, p := range q.GetClassProps(class, "$"+s.Name, nil).ReduceAccess(currentClass, s) {
				s.Type.merge(p.Prop.Types)
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
