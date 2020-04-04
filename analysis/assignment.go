package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
	"github.com/john-nguyen09/phpintel/util"
)

func newAssignment(document *Document, node *ast.Node) Symbol {
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if firstChild.Type() == "variable_name" {
		analyseVariableAssignment(document, firstChild, traverser.Clone(), node)
	} else {
		scanNode(document, firstChild)
		hasEqual := false
		child := traverser.Advance()
		for child != nil {
			if child.Type() == "=" {
				hasEqual = true
			}
			if hasEqual {
				scanNode(document, child)
			}
			child = traverser.Advance()
		}
	}
	return nil
}

func analyseVariableAssignment(document *Document, node *ast.Node, traverser *util.Traverser, parent *ast.Node) {
	traverser.Advance()
	traverser.SkipToken(" ")
	if parent.Type() == "augmented_assignment_expression" {
		traverser.SkipToken("operator")
	} else {
		traverser.SkipToken("=")
	}
	traverser.SkipToken(" ")
	traverser.SkipToken("&")
	rhs := traverser.Advance()
	phpDoc := document.getValidPhpDoc(document.GetNodeLocation(node))
	variable, shouldAdd := newVariable(document, node)
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
