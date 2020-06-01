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
	s.hasResolved = true
	q := ctx.query
	var scopeName string
	if n, ok := s.Scope.(HasName); ok {
		scopeName = n.GetName()
	}
	currentClass := ctx.document.GetClassScopeAtSymbol(s.Scope)
	for _, scopeType := range s.ResolveAndGetScope(ctx).Resolve() {
		for _, class := range q.GetClasses(scopeType.GetFQN()) {
			for _, m := range q.GetClassMethods(class, s.Name, nil).ReduceStatic(currentClass, scopeName) {
				s.Type.merge(resolveMemberTypes(m.Method.GetReturnTypes(), s.Scope))
			}
		}
		for _, intf := range q.GetInterfaces(scopeType.GetFQN()) {
			for _, m := range q.GetInterfaceMethods(intf, s.Name, nil).ReduceStatic(currentClass, scopeName) {
				s.Type.merge(resolveMemberTypes(m.Method.GetReturnTypes(), s.Scope))
			}
		}
	}
}

func (s *ScopedMethodAccess) GetTypes() TypeComposite {
	return s.Type
}

func (s *ScopedMethodAccess) ResolveToHasParams(ctx ResolveContext) []HasParams {
	hasParams := []HasParams{}
	q := ctx.query
	var scopeName string
	if n, ok := s.Scope.(HasName); ok {
		scopeName = n.GetName()
	}
	currentClass := ctx.document.GetClassScopeAtSymbol(s.Scope)
	for _, scopeType := range s.ResolveAndGetScope(ctx).Resolve() {
		for _, class := range q.GetClasses(scopeType.GetFQN()) {
			for _, m := range q.GetClassMethods(class, s.Name, nil).ReduceStatic(currentClass, scopeName) {
				hasParams = append(hasParams, m.Method)
			}
		}
		for _, intf := range q.GetInterfaces(scopeType.GetFQN()) {
			for _, m := range q.GetInterfaceMethods(intf, s.Name, nil).ReduceStatic(currentClass, scopeName) {
				hasParams = append(hasParams, m.Method)
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
