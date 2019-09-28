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

func newFunctionCall(document *Document, parent symbolBlock, node *phrase.Phrase) Symbol {
	functionCall := &FunctionCall{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	if len(node.Children) >= 1 {
		functionCall.Name = util.GetNodeText(node.Children[0], document.GetText())
	}
	nameLowerCase := strings.ToLower(functionCall.Name)
	if nameLowerCase == "\\define" || nameLowerCase == "define" {
		return newDefine(document, parent, node)
	}
	scanForChildren(parent, node)
	return functionCall
}

func (s *FunctionCall) getLocation() lsp.Location {
	return s.Location
}
