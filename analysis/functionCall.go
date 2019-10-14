package analysis

import (
	"strings"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// FunctionCall represents a reference to function call
type FunctionCall struct {
	Expression
}

func tryToNewDefine(document *Document, parent symbolBlock, node *phrase.Phrase) Symbol {
	if len(node.Children) >= 1 {
		nameLowerCase := strings.ToLower(util.GetNodeText(node.Children[0], document.GetText()))
		if nameLowerCase == "\\define" || nameLowerCase == "define" {
			return newDefine(document, parent, node)
		}
	}
	return nil
}

func newFunctionCall(document *Document, parent symbolBlock, node *phrase.Phrase) hasTypes {
	functionCall := &FunctionCall{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	if len(node.Children) >= 1 {
		functionCall.Name = util.GetNodeText(node.Children[0], document.GetText())
	}
	return functionCall
}

func (s *FunctionCall) getLocation() lsp.Location {
	return s.Location
}

func (s *FunctionCall) getTypes() TypeComposite {
	// TODO: Look up function for return types
	return s.Type
}
