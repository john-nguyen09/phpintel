package analysis

import (
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

func newAssignment(document *Document, node *sitter.Node) (HasTypes, bool) {
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

func analyseVariableAssignment(document *Document, lhs *sitter.Node, rhs *sitter.Node, parent *sitter.Node) {
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
	document.pushVariable(variable, util.PointToPosition(parent.EndPoint()))
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
