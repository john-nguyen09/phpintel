package analysis

import (
	"strings"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// FunctionCall represents a reference to function call
type FunctionCall struct {
	location lsp.Location

	ArgumentList ArgumentList
}

func newFunctionCall(document *Document, parent symbolBlock, node *phrase.Phrase) Symbol {
	if node.Type == phrase.FunctionCallExpression &&
		len(node.Children) >= 1 {
		text := strings.ToLower(util.GetNodeText(node.Children[0], document.GetText()))
		if text == "\\define" || text == "define" {
			return newDefine(document, parent, node)
		}
	}

	return nil
}
