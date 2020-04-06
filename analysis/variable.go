package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// Variable represents a reference to the variable
type Variable struct {
	Expression
	description        string
	canReferenceGlobal bool
	hasResolved        bool
}

var _ CanAddType = (*Variable)(nil)

func newVariableExpression(document *Document, node *ast.Node) (HasTypes, bool) {
	return newVariable(document, node)
}

func newVariable(document *Document, node *ast.Node) (*Variable, bool) {
	variable := newVariableWithoutPushing(document, node)
	document.pushVariable(variable)
	return variable, true
}

func newVariableWithoutPushing(document *Document, node *ast.Node) *Variable {
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

func (s *Variable) AddTypes(t TypeComposite) {
	s.Type.merge(t)
}

// VariableTable holds the range and the variables inside
type VariableTable struct {
	locationRange  protocol.Range
	variables      map[string]*Variable
	globalDeclares map[string]bool
	level          int
	children       []*VariableTable
}

func newVariableTable(locationRange protocol.Range, level int) *VariableTable {
	return &VariableTable{
		locationRange:  locationRange,
		variables:      map[string]*Variable{},
		globalDeclares: map[string]bool{},
		level:          level,
	}
}

func (vt *VariableTable) add(variable *Variable) {
	vt.variables[variable.Name] = variable
}

func (vt *VariableTable) get(name string) *Variable {
	if variable, ok := vt.variables[name]; ok {
		return variable
	}
	return nil
}

func (vt *VariableTable) canReferenceGlobal(name string) bool {
	if _, ok := vt.globalDeclares[name]; ok {
		return true
	}
	return false
}

func (vt *VariableTable) setReferenceGlobal(name string) {
	vt.globalDeclares[name] = true
}

// GetVariables returns all the variables in the table
func (vt *VariableTable) GetVariables() map[string]*Variable {
	return vt.variables
}

func (vt *VariableTable) addChild(child *VariableTable) {
	vt.children = append(vt.children, child)
}
