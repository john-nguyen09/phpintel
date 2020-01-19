package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

func newAssignment(document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		if p.Type == phrase.SimpleVariable {
			analyseVariableAssignment(document, p, traverser.Clone(), node)
		}
	}
	scanForChildren(document, node)
	return nil
}

func analyseVariableAssignment(document *Document, node *phrase.Phrase, traverser *util.Traverser, parent *phrase.Phrase) {
	traverser.Advance()
	traverser.SkipToken(lexer.Whitespace)
	traverser.SkipToken(lexer.Equals)
	traverser.SkipToken(lexer.Whitespace)
	if parent.Type == phrase.ByRefAssignmentExpression {
		traverser.SkipToken(lexer.Ampersand)
		traverser.SkipToken(lexer.Whitespace)
	}
	rhs := traverser.Advance()
	phpDoc := document.getValidPhpDoc(document.GetNodeLocation(node))
	variable, _ := newVariable(document, node)
	if phpDoc != nil {
		variable.applyPhpDoc(document, *phpDoc)
	}
	document.addSymbol(variable)

	var expression HasTypes = nil
	if p, ok := rhs.(*phrase.Phrase); ok {
		expression = scanForExpression(document, p)
	}
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
