package analysis

import (
	"encoding/json"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Variable represents a reference to the variable
type Variable struct {
	Expression
}

func newVariableExpression(document *Document, node *phrase.Phrase) hasTypes {
	return newVariable(document, node)
}

func newVariable(document *Document, node *phrase.Phrase) *Variable {
	variable := &Variable{
		Expression: Expression{
			Name:     util.GetNodeText(node, document.GetText()),
			Location: document.GetNodeLocation(node),
		},
	}
	document.pushVariable(variable)
	return variable
}

func (s *Variable) getLocation() lsp.Location {
	return s.Location
}

func (s *Variable) setExpression(expression hasTypes) {
	s.Scope = expression
}

func (s *Variable) mergeTypesWithVariable(variable *Variable) {
	types := variable.getTypes()
	for _, typeString := range types.Resolve() {
		s.Type.add(typeString)
	}
}

func (s *Variable) getTypes() TypeComposite {
	if s.Scope == nil {
		return s.Type
	}
	types := s.Scope.getTypes()
	for _, typeString := range s.Type.Resolve() {
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

func (s *Variable) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadVariable(serialiser *Serialiser) *Variable {
	return &Variable{
		Expression: ReadExpression(serialiser),
	}
}
