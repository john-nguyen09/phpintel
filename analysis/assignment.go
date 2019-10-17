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
			analyseVariableAssignment(document, p, traverser.Clone())
		}
	}
	return nil
}

func analyseVariableAssignment(document *Document,
	node *phrase.Phrase, traverser util.Traverser) {
	traverser.Advance()
	traverser.SkipToken(lexer.Whitespace)
	traverser.SkipToken(lexer.Equals)
	traverser.SkipToken(lexer.Whitespace)
	rhs := traverser.Advance()
	variable := newVariable(document, node)
	document.addSymbol(variable)

	var expression hasTypes = nil
	if p, ok := rhs.(*phrase.Phrase); ok {
		expression = scanForExpression(document, p)
	}
	if expression != nil {
		variable.setExpression(expression)
	}
}
