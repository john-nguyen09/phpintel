package analysis

import (
	"strings"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// FunctionCall represents a reference to function call
type FunctionCall struct {
	Expression
}

func tryToNewDefine(document *Document, node *phrase.Phrase) Symbol {
	if len(node.Children) >= 1 {
		nameLowerCase := strings.ToLower(document.GetNodeText(node.Children[0]))
		if nameLowerCase == "\\define" || nameLowerCase == "define" {
			return newDefine(document, node)
		}
		scanForChildren(document, node)
	}
	return nil
}

func newFunctionCall(document *Document, node *phrase.Phrase) HasTypes {
	functionCall := &FunctionCall{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	if len(node.Children) >= 1 {
		functionCall.Name = document.GetNodeText(node.Children[0])
	}
	return functionCall
}

func (s *FunctionCall) GetLocation() protocol.Location {
	return s.Location
}

func (s *FunctionCall) GetTypes() TypeComposite {
	// TODO: Look up function for return types
	return s.Type
}

func (s *FunctionCall) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadFunctionCall(serialiser *Serialiser) *FunctionCall {
	return &FunctionCall{
		Expression: ReadExpression(serialiser),
	}
}
