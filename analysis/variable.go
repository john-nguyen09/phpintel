package analysis

import (
	"encoding/json"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Variable represents a reference to the variable
type Variable struct {
	Name       string
	location   lsp.Location
	types      TypeComposite
	expression hasTypes
}

func newVariableExpression(document *Document, parent symbolBlock, node *phrase.Phrase) hasTypes {
	return newVariable(document, parent, node)
}

func newVariable(document *Document, parent symbolBlock, node *phrase.Phrase) *Variable {
	variable := &Variable{
		Name:     util.GetNodeText(node, document.GetText()),
		location: document.GetNodeLocation(node),
	}
	document.pushVariable(variable)
	return variable
}

func (s *Variable) getLocation() lsp.Location {
	return s.location
}

func (s *Variable) setExpression(expression hasTypes) {
	s.expression = expression
}

func (s *Variable) mergeTypesWithVariable(variable *Variable) {
	types := variable.getTypes()
	for _, typeString := range types.Resolve() {
		s.types.add(typeString)
	}
}

func (s *Variable) getTypes() TypeComposite {
	if s.expression == nil {
		return s.types
	}
	types := s.expression.getTypes()
	for _, typeString := range s.types.Resolve() {
		types.add(typeString)
	}
	return types
}

func (s *Variable) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name     string
		Location lsp.Location
		Types    TypeComposite
	}{
		Name:     s.Name,
		Location: s.getLocation(),
		Types:    s.getTypes(),
	})
}
