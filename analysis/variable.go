package analysis

import (
	"encoding/json"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Variable represents a reference to the variable
type Variable struct {
	Expression
	description        string
	canReferenceGlobal bool
	hasResolved        bool
}

func newVariableExpression(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	return newVariable(document, node)
}

func newVariable(document *Document, node *phrase.Phrase) (*Variable, bool) {
	variable := &Variable{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if t, ok := child.(*lexer.Token); ok {
			if t.Type == lexer.VariableName {
				variable.Name = document.GetTokenText(t)
			}
		}
		child = traverser.Advance()
	}
	if variable.Name == "$this" {
		variable.setExpression(newRelativeScope(document, variable.Location))
	}
	document.pushVariable(variable)
	return variable, true
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

func (s *Variable) Resolve(store *Store) {
	if s.hasResolved {
		return
	}
	s.hasResolved = true
	if !s.canReferenceGlobal {
		return
	}
	globalVariables := store.GetGlobalVariables(s.Name)
	for _, globalVariable := range globalVariables {
		s.Type.merge(globalVariable.types)
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

func (s *Variable) GetDetail() string {
	return s.Type.ToString()
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
