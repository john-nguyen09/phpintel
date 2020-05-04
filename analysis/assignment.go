package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

func newAssignment(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	isVariable := false
	if p, ok := firstChild.(*phrase.Phrase); ok {
		if p.Type == phrase.SimpleVariable {
			analyseVariableAssignment(document, p, traverser.Clone(), node)
			isVariable = true
		}
	}
	if !isVariable {
		if p, ok := firstChild.(*phrase.Phrase); ok {
			scanNode(document, p)
		}
		for child := traverser.Advance(); child != nil; child = traverser.Advance() {
			scanNode(document, child)
		}
	}
	return nil, false
}

func analyseVariableAssignment(document *Document, lhs *phrase.Phrase, traverser *util.Traverser, parent *phrase.Phrase) {
	traverser.Advance()
	traverser.SkipToken(lexer.Whitespace)
	if parent.Type == phrase.CompoundAssignmentExpression {
		traverser.SkipToken(lexer.DotEquals)
	} else {
		traverser.SkipToken(lexer.Equals)
	}
	traverser.SkipToken(lexer.Whitespace)
	if parent.Type == phrase.ByRefAssignmentExpression {
		traverser.SkipToken(lexer.Ampersand)
		traverser.SkipToken(lexer.Whitespace)
	}
	rhs := traverser.Advance()
	phpDoc := document.getValidPhpDoc(document.GetNodeLocation(lhs))
	variable := newVariableWithoutPushing(document, lhs)
	if phpDoc != nil {
		variable.applyPhpDoc(document, *phpDoc)
	}
	// The variable appears before rhs therefore being added before rhs
	document.addSymbol(variable)

	var expression HasTypes = nil
	if p, ok := rhs.(*phrase.Phrase); ok {
		expression = scanForExpression(document, p)
	}
	// But the variable should be pushed after any rhs's variables
	lastToken := util.LastToken(parent)
	document.pushVariable(variable, document.positionAt(lastToken.Offset+lastToken.Length))
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
