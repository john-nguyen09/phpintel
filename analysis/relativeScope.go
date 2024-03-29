package analysis

import "github.com/john-nguyen09/phpintel/internal/lsp/protocol"

type RelativeScope struct {
	location protocol.Location
	Types    TypeComposite
}

func IsNameRelative(name string) bool {
	return name == "static" || name == "self" || name == "$this"
}

func IsNameParent(name string) bool {
	return name == "parent"
}

func resolveMemberTypes(types TypeComposite, scope HasTypes) TypeComposite {
	newTypes := newTypeComposite()
	for _, t := range types.Resolve() {
		if t.GetOriginal() == "static" || t.GetOriginal() == "$this" {
			newTypes.mergeWithArrayLevel(scope.GetTypes(), t.arrayLevel)
			continue
		}
		newTypes.add(t)
	}
	return newTypes
}

func newRelativeScope(document *Document, location protocol.Location) *RelativeScope {
	types := newTypeComposite()
	lastClass := document.getLastClass()
	switch v := lastClass.(type) {
	case *Class:
		types.add(v.Name)
	case *Interface:
		types.add(v.Name)
	case *Trait:
		types.add(v.Name)
	}
	return &RelativeScope{
		location: location,
		Types:    types,
	}
}

func newParentScope(document *Document, location protocol.Location) *RelativeScope {
	types := newTypeComposite()
	lastClass := document.getLastClass()
	switch v := lastClass.(type) {
	case *Class:
		types.add(v.Extends)
	case *Interface:
		for _, extend := range v.Extends {
			types.add(extend)
		}
	}
	return &RelativeScope{
		location: location,
		Types:    types,
	}
}

func (s *RelativeScope) GetLocation() protocol.Location {
	return s.location
}

func (s *RelativeScope) GetTypes() TypeComposite {
	return s.Types
}

func (s *RelativeScope) Resolve(ctx ResolveContext) {

}
