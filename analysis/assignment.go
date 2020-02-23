package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/john-nguyen09/phpintel/util"
)

func newAssignment(document *Document, node *sitter.Node) Symbol {
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if firstChild.Type() == "variable" {
		analyseVariableAssignment(document, firstChild, traverser.Clone(), node)
	}
	scanForChildren(document, node)
	return nil
}

func analyseVariableAssignment(document *Document, node *sitter.Node, traverser *util.Traverser, parent *sitter.Node) {
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
	variable, _ := newVariable(document, node)
	if phpDoc != nil {
		variable.applyPhpDoc(document, *phpDoc)
	}
	document.addSymbol(variable)

	var expression HasTypes = nil
	expression = scanForExpression(document, rhs)
	if expression != nil {
		variable.setExpression(expression)
	}
	globalVariable := document.getGlobalVariable(variable.Name)
	if globalVariable != nil {
		types := variable.GetTypes()
		if !types.IsEmpty() {
			globalVariable.types.merge(types)
		}
	}
}
