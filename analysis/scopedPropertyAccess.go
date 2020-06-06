package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ScopedPropertyAccess represents a reference to property in
// scoped class access, e.g. ::$property
type ScopedPropertyAccess struct {
	MemberAccessExpression

	hasResolved bool
}

var _ HasTypesHasScope = (*ScopedPropertyAccess)(nil)
var _ MemberAccess = (*ScopedPropertyAccess)(nil)

func newScopedPropertyAccess(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	propertyAccess := &ScopedPropertyAccess{
		MemberAccessExpression: MemberAccessExpression{
			Expression: Expression{},
		},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		classAccess := newClassAccess(a, document, p)
		document.addSymbol(classAccess)
		propertyAccess.Scope = classAccess
	}
	traverser.Advance()
	thirdChild := traverser.Advance()
	propertyAccess.Location = document.GetNodeLocation(thirdChild)
	if p, ok := thirdChild.(*phrase.Phrase); ok {
		propertyAccess.Name = analyseMemberName(document, p)
	}
	return propertyAccess, true
}

func (s *ScopedPropertyAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ScopedPropertyAccess) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	q := ctx.query
	s.hasResolved = true
	currentClass := ctx.document.GetClassScopeAtSymbol(s.Scope)
	for _, scopeType := range s.ResolveAndGetScope(ctx).Resolve() {
		for _, class := range q.GetClasses(scopeType.GetFQN()) {
			for _, p := range q.GetClassProps(class, s.Name, nil).ReduceStatic(currentClass, s) {
				s.Type.merge(p.Prop.Types)
			}
		}
	}
}

func (s *ScopedPropertyAccess) GetTypes() TypeComposite {
	return s.Type
}

func (s *ScopedPropertyAccess) GetScopeTypes() TypeComposite {
	if s.Scope != nil {
		return s.Scope.GetTypes()
	}
	return newTypeComposite()
}

func (s *ScopedPropertyAccess) MemberName() string {
	return s.Name
}
