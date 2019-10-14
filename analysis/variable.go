package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Variable represents a reference to the variable
type Variable struct {
	Name       string
	location   lsp.Location
	expression hasTypes
}

func newVariableExpression(document *Document, parent symbolBlock, node *phrase.Phrase) hasTypes {
	return newVariable(document, parent, node)
}

func newVariable(document *Document, parent symbolBlock, node *phrase.Phrase) *Variable {
	variable := &Variable{
		Name:     util.GetNodeText(node, document.GetText()),
		location: document.GetNodeLocation(node),
	}
	return variable
}

func (s *Variable) getLocation() lsp.Location {
	return s.location
}

func (s *Variable) setExpression(expression hasTypes) {
	s.expression = expression
}

func (s *Variable) getTypes() TypeComposite {
	return s.expression.getTypes()
}
