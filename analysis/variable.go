package analysis

import (
	"encoding/json"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	sitter "github.com/smacker/go-tree-sitter"
)

// Variable represents a reference to the variable
type Variable struct {
	Expression
	description        string
	canReferenceGlobal bool
	hasResolved        bool
}

func newVariableExpression(document *Document, node *sitter.Node) (HasTypes, bool) {
	return newVariable(document, node)
}

func newVariable(document *Document, node *sitter.Node) (*Variable, bool) {
	variable := &Variable{
		Expression: Expression{
			Name:     document.GetNodeText(node),
			Location: document.GetNodeLocation(node),
		},
	}
	phpDoc := document.getValidPhpDoc(variable.Location)
	if phpDoc != nil {
		variable.applyPhpDoc(document, *phpDoc)
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
	if s.Scope == nil {
		s.setExpression(variable.Scope)
	}
}

func (s *Variable) applyPhpDoc(document *Document, phpDoc phpDocComment) {
	for _, varTag := range phpDoc.Vars {
		s.Type.merge(typesFromPhpDoc(document, varTag.TypeString))
	}
}

func (s *Variable) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	store := ctx.store
	s.hasResolved = true
	if s.canReferenceGlobal {
		globalVariables := store.GetGlobalVariables(s.Name)
		for _, globalVariable := range globalVariables {
			s.Type.merge(globalVariable.types)
		}
	}
	if s.Scope != nil {
		s.Scope.Resolve(ctx)
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

func (s *Variable) GetName() string {
	return s.Name
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
