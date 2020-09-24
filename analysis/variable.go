package analysis

import (
	"sort"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// Variable represents a reference to the variable
type Variable struct {
	Expression
	description        string
	canReferenceGlobal bool
	hasResolved        bool
}

func newVariableExpression(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	return newVariable(a, document, node, false)
}

func newVariable(a analyser, document *Document, node *phrase.Phrase, isDeclaration bool) (*Variable, bool) {
	variable := newVariableWithoutPushing(document, node)
	document.pushVariable(a, variable, variable.GetLocation().Range.End, isDeclaration)
	return variable, true
}

func newVariableWithoutPushing(document *Document, node *phrase.Phrase) *Variable {
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
	q := ctx.query
	s.hasResolved = true
	if s.canReferenceGlobal {
		globalVariables := q.GetGlobalVariables(s.Name)
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
	return s.GetTypes().ToString()
}

func (s *Variable) GetName() string {
	return s.Name
}

type contextualVariable struct {
	v             *Variable
	start         protocol.Position
	isDeclaration bool
	// If isDeclaration is false then this is not used
	isUsed bool
}

func newContextualVariable(v *Variable, start protocol.Position, isDeclaration bool) contextualVariable {
	return contextualVariable{
		v:             v,
		start:         start,
		isDeclaration: isDeclaration,
	}
}

// Variable returns the variable of the contextualVariable
func (c contextualVariable) Variable() *Variable {
	return c.v
}

// VariableTable holds the range and the variables inside
type VariableTable struct {
	locationRange  protocol.Range
	variables      map[string][]contextualVariable
	globalDeclares map[string]bool
	level          int
	children       []*VariableTable
}

func newVariableTable(locationRange protocol.Range, level int) *VariableTable {
	return &VariableTable{
		locationRange:  locationRange,
		variables:      map[string][]contextualVariable{},
		globalDeclares: map[string]bool{},
		level:          level,
	}
}

func (vt *VariableTable) add(a analyser, variable *Variable, start protocol.Position, isDeclaration bool) {
	currentVars := []contextualVariable{}
	newCtxVar := newContextualVariable(variable, start, isDeclaration)
	if !isDeclaration {
		pos := variable.GetLocation().Range.Start
		if ctxVars, ok := vt.variables[variable.Name]; ok {
			for i, ctxVar := range ctxVars {
				if protocol.ComparePos(pos, ctxVar.v.GetLocation().Range.Start) <= 0 {
					break
				}
				if ctxVar.isDeclaration && ctxVar.v.Name == variable.Name {
					ctxVars[i].isUsed = true
				}
			}
		}
	}
	if prevVars, ok := vt.variables[variable.Name]; ok {
		index := 0
		if len(prevVars) > 0 {
			index = sort.Search(len(prevVars), func(i int) bool {
				return protocol.ComparePos(start, prevVars[i].start) < 0
			})
		}
		prevVars = append(prevVars[:index], append([]contextualVariable{newCtxVar}, prevVars[index:]...)...)
		vt.variables[variable.Name] = prevVars
	} else {
		vt.variables[variable.Name] = append(currentVars, newCtxVar)
	}
}

func (vt *VariableTable) get(name string, pos protocol.Position) *Variable {
	if vars, ok := vt.variables[name]; ok {
		index := sort.Search(len(vars), func(i int) bool {
			return protocol.ComparePos(pos, vars[i].start) < 0
		})
		index--
		if index >= 0 && index < len(vars) {
			return vars[index].v
		}
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
func (vt *VariableTable) GetVariables(pos protocol.Position) []*Variable {
	results := []*Variable{}
	for _, vars := range vt.variables {
		if len(vars) == 0 {
			continue
		}
		index := sort.Search(len(vars), func(i int) bool {
			return protocol.ComparePos(vars[i].start, pos) >= 0
		})
		index--
		if index >= 0 && index < len(vars) {
			results = append(results, vars[index].v)
		}
	}
	return results
}

func (vt *VariableTable) addChild(child *VariableTable) {
	vt.children = append(vt.children, child)
}

func (vt *VariableTable) unusedVariables() []*Variable {
	results := []*Variable{}
	for _, ctxVars := range vt.variables {
		for _, ctxVar := range ctxVars {
			if ctxVar.isDeclaration && !ctxVar.isUsed {
				results = append(results, ctxVar.v)
			}
		}
	}
	return results
}

// GetContextualVariables returns the variable instances with the given name
func (vt *VariableTable) GetContextualVariables(name string) []contextualVariable {
	if ctxVars, ok := vt.variables[name]; ok {
		return ctxVars
	}
	return nil
}
