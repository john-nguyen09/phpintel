package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/indexer"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Parameter contains information of a function parameter
type Parameter struct {
	location lsp.Location

	Name  string        `json:"Name"`
	Type  TypeComposite `json:"Type"`
	Value string        `json:"Value"`
}

func newParameter(document *Document, node *phrase.Phrase) *Parameter {
	param := &Parameter{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	hasEqual := false
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.TypeDeclaration:
				{
					typeDeclaration := newTypeDeclaration(document, p)
					for _, typeString := range typeDeclaration.Type.typeStrings {
						param.Type.add(typeString)
					}
				}
			}

			if hasEqual {
				param.Value += util.GetNodeText(p, document.GetText())
			}
		} else if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Equals:
				{
					hasEqual = true
				}
			case lexer.VariableName:
				{
					param.Name = util.GetNodeText(token, document.GetText())
				}
			}
		}
		child = traverser.Advance()
	}

	return param
}

func (s *Parameter) Write(serialiser *indexer.Serialiser) {
	util.WriteLocation(serialiser, s.location)
	serialiser.WriteString(s.Name)
	s.Type.Write(serialiser)
	serialiser.WriteString(s.Value)
}

func ReadParameter(serialiser *indexer.Serialiser) Parameter {
	return Parameter{
		location: util.ReadLocation(serialiser),
		Name:     serialiser.ReadString(),
		Type:     ReadTypeComposite(serialiser),
		Value:    serialiser.ReadString(),
	}
}
