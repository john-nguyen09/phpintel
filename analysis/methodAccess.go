package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type MethodAccess struct {
	MemberAccessExpression

	hasResolved bool
}

var _ HasTypesHasScope = (*MethodAccess)(nil)

func newMethodAccess(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	methodAccess := &MethodAccess{
		MemberAccessExpression: MemberAccessExpression{
			Expression: Expression{},
		},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		expression := scanForExpression(a, document, p)
		if expression != nil {
			methodAccess.Scope = expression
		}
	}
	traverser.Advance()
	methodAccess.Name, methodAccess.Location = readMemberName(a, document, traverser)
	document.addSymbol(methodAccess)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ArgumentExpressionList:
				newArgumentList(a, document, p)
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
	q := ctx.query
	currentClass := ctx.document.GetClassScopeAtSymbol(s.Scope)
	for _, scopeType := range s.ResolveAndGetScope(ctx).Resolve() {
		for _, class := range q.GetClasses(scopeType.GetFQN()) {
			for _, m := range q.GetClassMethods(class, s.Name, nil).ReduceAccess(currentClass, s) {
				s.Type.merge(resolveMemberTypes(m.Method.GetReturnTypes(), s.Scope))
			}
		}
		for _, theInterface := range q.GetInterfaces(scopeType.GetFQN()) {
			for _, m := range q.GetInterfaceMethods(theInterface, s.Name, nil).ReduceAccess(currentClass, s) {
				s.Type.merge(resolveMemberTypes(m.Method.GetReturnTypes(), s.Scope))
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
	q := ctx.query
	currentClass := ctx.document.GetClassScopeAtSymbol(s.Scope)
	types := s.ResolveAndGetScope(ctx)
	for _, typeString := range types.Resolve() {
		methods := []MethodWithScope{}
		for _, class := range q.GetClasses(typeString.GetFQN()) {
			methods = append(methods, q.GetClassMethods(class, s.Name, nil).ReduceAccess(currentClass, s)...)
		}
		for _, intf := range q.GetInterfaces(typeString.GetFQN()) {
			methods = append(methods, q.GetInterfaceMethods(intf, s.Name, nil).ReduceAccess(currentClass, s)...)
		}
		for _, m := range methods {
			hasParams = append(hasParams, m.Method)
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
