package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// ScopedMethodAccess represents a reference to method in class access, e.g. ::method()
type ScopedMethodAccess struct {
	Expression

	hasResolved bool
}

func newScopedMethodAccess(document *Document, node *sitter.Node) (HasTypes, bool) {
	methodAccess := &ScopedMethodAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	classAccess := newClassAccess(document, firstChild)
	document.addSymbol(classAccess)
	methodAccess.Scope = classAccess
	traverser.Advance()
	thirdChild := traverser.Advance()
	methodAccess.Location = document.GetNodeLocation(thirdChild)
	methodAccess.Name = analyseMemberName(document, thirdChild)
	document.addSymbol(methodAccess)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "arguments":
			scanNode(document, child)
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
				s.Type.merge(method.GetReturnTypes())
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
