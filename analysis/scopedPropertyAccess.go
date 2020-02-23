package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// ScopedPropertyAccess represents a reference to property in
// scoped class access, e.g. ::$property
type ScopedPropertyAccess struct {
	Expression

	hasResolved bool
}

func newScopedPropertyAccess(document *Document, node *sitter.Node) (HasTypes, bool) {
	propertyAccess := &ScopedPropertyAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	classAccess := newClassAccess(document, firstChild)
	document.addSymbol(classAccess)
	propertyAccess.Scope = classAccess
	traverser.Advance()
	thirdChild := traverser.Advance()
	propertyAccess.Location = document.GetNodeLocation(thirdChild)
	propertyAccess.Name = analyseMemberName(document, thirdChild)
	return propertyAccess, true
}

func (s *ScopedPropertyAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ScopedPropertyAccess) Resolve(ctx ResolveContext) {
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
			for _, property := range GetClassProperties(store, class, s.Name,
				StaticPropsScopeAware(NewSearchOptions(), classScope, name)) {
				s.Type.merge(property.Types)
			}
		}
	}
}

func (s *ScopedPropertyAccess) GetTypes() TypeComposite {
	return s.Type
}

func (s *ScopedPropertyAccess) Serialise(e *storage.Encoder) {
	s.Expression.Serialise(e)
}

func ReadScopedPropertyAccess(d *storage.Decoder) *ScopedPropertyAccess {
	return &ScopedPropertyAccess{
		Expression: ReadExpression(d),
	}
}
