package analysis

import (
	"strings"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type FunctionCall struct {
	location lsp.Location

	ArgumentList ArgumentList
}

func NewFunctionCall(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	if node.Type == phrase.FunctionCallExpression &&
		len(node.Children) >= 1 {
		text := strings.ToLower(util.GetNodeText(node.Children[0], document.GetText()))
		if text == "\\define" || text == "define" {
			return NewDefine(document, parent, node)
		}
	}

	return nil
}
