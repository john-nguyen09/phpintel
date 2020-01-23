package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ScopedPropertyAccess represents a reference to property in
// scoped class access, e.g. ::$property
type ScopedPropertyAccess struct {
	Expression

	hasResolved bool
}

func newScopedPropertyAccess(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	propertyAccess := &ScopedPropertyAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		classAccess := newClassAccess(document, p)
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

func (s *ScopedPropertyAccess) Resolve(store *Store) {
	if s.hasResolved {
		return
	}
	s.hasResolved = true
	name := ""
	classScope := ""
	if hasName, ok := s.Scope.(HasName); ok {
		name = hasName.GetName()
	}
	if hasScope, ok := s.Scope.(HasScope); ok {
		classScope = hasScope.GetScope().GetFQN()
	}
	for _, scopeType := range s.ResolveAndGetScope(store).Resolve() {
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

func (s *ScopedPropertyAccess) Serialiser(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadScopedPropertyAccess(serialiser *Serialiser) *ScopedPropertyAccess {
	return &ScopedPropertyAccess{
		Expression: ReadExpression(serialiser),
	}
}
