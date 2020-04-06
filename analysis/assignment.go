package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
)

func newAssignment(document *Document, node *ast.Node) (HasTypes, bool) {
	lhs := node.ChildByFieldName("left")
	rhs := node.ChildByFieldName("right")
	if lhs != nil {
		if lhs.Type() == "variable_name" {
			analyseVariableAssignment(document, lhs, rhs, node)
		} else {
			scanNode(document, lhs)
			if rhs != nil {
				scanNode(document, rhs)
			}
		}
	}
	return nil, false
}

func analyseVariableAssignment(document *Document, lhs *ast.Node, rhs *ast.Node, parent *ast.Node) {
	phpDoc := document.getValidPhpDoc(document.GetNodeLocation(parent))
	variable := newVariableWithoutPushing(document, lhs)
	if phpDoc != nil {
		variable.applyPhpDoc(document, *phpDoc)
	}
	// The variable appears before rhs therefore being added before rhs
	document.addSymbol(variable)

	var expression HasTypes = nil
	expression = scanForExpression(document, rhs)
	// But the variable should be pushed after any rhs's variables
	document.pushVariable(variable)
	if expression != nil {
		variable.setExpression(expression)
	} else {
		scanNode(document, rhs)
	}
	globalVariable := document.getGlobalVariable(variable.Name)
	if globalVariable != nil {
		types := variable.GetTypes()
		if !types.IsEmpty() {
			globalVariable.types.merge(types)
		}
	}
}
