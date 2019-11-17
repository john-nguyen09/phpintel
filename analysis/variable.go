package analysis

import (
	"encoding/json"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// Variable represents a reference to the variable
type Variable struct {
	Expression
	description string
}

func newVariableExpression(document *Document, node *phrase.Phrase) HasTypes {
	return newVariable(document, node)
}

func newVariable(document *Document, node *phrase.Phrase) *Variable {
	variable := &Variable{
		Expression: Expression{
			Name:     document.GetNodeText(node),
			Location: document.GetNodeLocation(node),
		},
	}
	document.pushVariable(variable)
	return variable
}

func (s *Variable) GetLocation() protocol.Location {
	return s.Location
}

func (s *Variable) setExpression(expression HasTypes) {
	s.Scope = expression
}

func (s *Variable) mergeTypesWithVariable(variable *Variable) {
	types := variable.GetTypes()
	for _, typeString := range types.Resolve() {
		s.Type.add(typeString)
	}
}

func (s *Variable) GetTypes() TypeComposite {
	if s.Scope == nil {
		return s.Type
	}
	types := s.Scope.GetTypes()
	for _, typeString := range s.Type.Resolve() {
		types.add(typeString)
	}
	return types
}

func (s *Variable) GetDescription() string {
	return s.description
}

func (s *Variable) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name     string
		Location protocol.Location
		Types    TypeComposite
	}{
		Name:     s.Name,
		Location: s.GetLocation(),
		Types:    s.GetTypes(),
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
