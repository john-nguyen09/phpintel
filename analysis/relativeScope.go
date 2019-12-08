package analysis

import "github.com/john-nguyen09/phpintel/internal/lsp/protocol"

type RelativeScope struct {
	location protocol.Location
	Types    TypeComposite
}

func isNameRelative(name string) bool {
	return name == "static" || name == "self"
}

func newRelativeScope(document *Document, location protocol.Location) *RelativeScope {
	types := newTypeComposite()
	lastClass := document.getLastClass()
	if class, ok := lastClass.(*Class); ok {
		types.add(class.Name)
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

func (s *RelativeScope) Resolve(store *Store) {

}
