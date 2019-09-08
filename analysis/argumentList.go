package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type ArgumentList struct {
	location lsp.Location

	arguments []phrase.AstNode
}

func NewArgumentList(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	argumentList := &ArgumentList{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			if token.Type == lexer.Whitespace || token.Type == lexer.Comma {
				child = traverser.Advance()
				continue
			}
		}
		argumentList.arguments = append(argumentList.arguments, child)
		child = traverser.Advance()
	}

	return argumentList
}

func (s *ArgumentList) GetLocation() lsp.Location {
	return s.location
}

func (s *ArgumentList) GetArguments() []phrase.AstNode {
	return s.arguments
}
