package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ScopedMethodAccess represents a reference to method in class access, e.g. ::method()
type ScopedMethodAccess struct {
	Expression

	hasResolved bool
}

var _ HasTypesHasScope = (*ScopedMethodAccess)(nil)

func newScopedMethodAccess(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	methodAccess := &ScopedMethodAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		classAccess := newClassAccess(a, document, p)
		document.addSymbol(classAccess)
		methodAccess.Scope = classAccess
	}
	traverser.Advance()
	thirdChild := traverser.Advance()
	methodAccess.Location = document.GetNodeLocation(thirdChild)
	if p, ok := thirdChild.(*phrase.Phrase); ok {
		methodAccess.Name = analyseMemberName(document, p)
	}
	document.addSymbol(methodAccess)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ArgumentExpressionList {
			scanNode(a, document, child)
			break
		}
		child = traverser.Advance()
	}
	return methodAccess, false
}

func (s *ScopedMethodAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ScopedMethodAccess) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	store := ctx.store
	s.hasResolved = true
	name := ""
	classScope := ""
	if hasName, ok := s.Scope.(HasName); ok {
		name = hasName.GetName()
	}
	if hasScope, ok := s.Scope.(HasScope); ok {
		classScope = hasScope.GetScope()
	}
	for _, scopeType := range s.ResolveAndGetScope(ctx).Resolve() {
		for _, class := range store.GetClasses(scopeType.GetFQN()) {
			for _, method := range GetClassMethods(store, class, s.Name,
				StaticMethodsScopeAware(NewSearchOptions(), classScope, name)) {
				s.Type.merge(resolveMemberTypes(method.GetReturnTypes(), s.Scope))
			}
		}
		for _, theInterface := range store.GetInterfaces(scopeType.GetFQN()) {
			for _, method := range GetInterfaceMethods(store, theInterface, s.Name,
				StaticMethodsScopeAware(NewSearchOptions(), classScope, name)) {
				s.Type.merge(resolveMemberTypes(method.GetReturnTypes(), s.Scope))
			}
		}
	}
}

func (s *ScopedMethodAccess) GetTypes() TypeComposite {
	return s.Type
}

func (s *ScopedMethodAccess) ResolveToHasParams(ctx ResolveContext) []HasParams {
	hasParams := []HasParams{}
	store := ctx.store
	for _, typeString := range s.ResolveAndGetScope(ctx).Resolve() {
		name := ""
		classScope := ""
		if hasName, ok := s.Scope.(HasName); ok {
			name = hasName.GetName()
		}
		if hasScope, ok := s.Scope.(HasScope); ok {
			classScope = hasScope.GetScope()
		}
		for _, class := range store.GetClasses(typeString.GetFQN()) {
			for _, method := range GetClassMethods(store, class, s.Name,
				StaticMethodsScopeAware(NewSearchOptions(), classScope, name)) {
				hasParams = append(hasParams, method)
			}
		}
	}
	return hasParams
}

func (s *ScopedMethodAccess) MemberName() string {
	return s.Name + "()"
}

func (s *ScopedMethodAccess) GetScopeTypes() TypeComposite {
	if s.Scope != nil {
		return s.Scope.GetTypes()
	}
	return newTypeComposite()
}
