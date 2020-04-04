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
	variable, shouldAdd := newVariable(document, lhs)
	if phpDoc != nil {
		variable.applyPhpDoc(document, *phpDoc)
	}
	if shouldAdd {
		document.addSymbol(variable)
	}

	var expression HasTypes = nil
	expression = scanForExpression(document, rhs)
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
