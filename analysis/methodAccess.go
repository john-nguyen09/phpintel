package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type MethodAccess struct {
	Expression

	hasResolved bool
}

var _ HasTypesHasScope = (*MethodAccess)(nil)

func newMethodAccess(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	methodAccess := &MethodAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		expression := scanForExpression(document, p)
		if expression != nil {
			methodAccess.Scope = expression
		}
	}
	traverser.Advance()
	methodAccess.Name, methodAccess.Location = readMemberName(document, traverser)
	document.addSymbol(methodAccess)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ArgumentExpressionList:
				newArgumentList(document, p)
				break
			}
		}
		child = traverser.Advance()
	}
	return methodAccess, false
}

func (s *MethodAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *MethodAccess) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	store := ctx.store
	for _, scopeType := range s.ResolveAndGetScope(ctx).Resolve() {
		for _, class := range store.GetClasses(scopeType.GetFQN()) {
			for _, method := range GetClassMethods(store, class, s.Name, NewSearchOptions()) {
				s.Type.merge(resolveMemberTypes(method.GetReturnTypes(), s.Scope))
			}
		}
		for _, theInterface := range store.GetInterfaces(scopeType.GetFQN()) {
			for _, method := range GetInterfaceMethods(store, theInterface, s.Name, NewSearchOptions()) {
				s.Type.merge(resolveMemberTypes(method.GetReturnTypes(), s.Scope))
			}
		}
	}
	s.hasResolved = true
}

func (s *MethodAccess) GetTypes() TypeComposite {
	return s.Type
}

func (s *MethodAccess) ResolveToHasParams(ctx ResolveContext) []HasParams {
	hasParams := []HasParams{}
	store := ctx.store
	document := ctx.document
	for _, typeString := range s.ResolveAndGetScope(ctx).Resolve() {
		methods := []*Method{}
		for _, class := range store.GetClasses(typeString.GetFQN()) {
			methods = append(methods, GetClassMethods(store, class, s.Name,
				MethodsScopeAware(NewSearchOptions(), document, s.Scope))...)
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			methods = append(methods, GetInterfaceMethods(store, theInterface, s.Name,
				MethodsScopeAware(NewSearchOptions(), document, s.Scope))...)
		}
		for _, trait := range store.GetTraits(typeString.GetFQN()) {
			methods = append(methods, GetTraitMethods(store, trait, s.Name,
				MethodsScopeAware(NewSearchOptions(), document, s.Scope))...)
		}
		for _, method := range methods {
			hasParams = append(hasParams, method)
		}
	}
	return hasParams
}

func (s *MethodAccess) GetScopeTypes() TypeComposite {
	if s.Scope != nil {
		return s.Scope.GetTypes()
	}
	return newTypeComposite()
}

func (s *MethodAccess) MemberName() string {
	return s.Name + "()"
}
