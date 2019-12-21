package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ArgumentList contains information of arguments in function-like call
type ArgumentList struct {
	location protocol.Location

	arguments []phrase.AstNode
	ranges    []protocol.Range
}

func newEmptyArgumentList(document *Document, open *lexer.Token, close *lexer.Token) *ArgumentList {
	closePos := document.positionAt(open.Offset)
	if close != nil {
		closePos = document.positionAt(close.Offset)
	}
	argumentList := &ArgumentList{
		location: protocol.Location{
			URI: document.GetURI(),
			Range: protocol.Range{
				Start: document.positionAt(open.Offset),
				End:   closePos,
			},
		},
	}
	argumentList.ranges = append(argumentList.ranges, argumentList.location.Range)
	return argumentList
}

func newArgumentList(document *Document, node *phrase.Phrase) Symbol {
	argumentList := &ArgumentList{
		location: document.GetNodeLocation(node),
	}
	document.addSymbol(argumentList)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	start := argumentList.location.Range.Start
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			if token.Type == lexer.Whitespace || token.Type == lexer.Comma {
				if token.Type == lexer.Comma {
					end := document.positionAt(token.Offset)
					argumentList.ranges = append(argumentList.ranges, protocol.Range{
						Start: start,
						End:   end,
					})
					start = end
				}
				child = traverser.Advance()
				continue
			}
		}
		argumentList.arguments = append(argumentList.arguments, child)
		child = traverser.Advance()
	}
	argumentList.ranges = append(argumentList.ranges, protocol.Range{
		Start: start,
		End:   argumentList.location.Range.End,
	})
	scanForChildren(document, node)
	return argumentList
}

func (s *ArgumentList) GetLocation() protocol.Location {
	return s.location
}

// GetArguments returns the arguments
func (s *ArgumentList) GetArguments() []phrase.AstNode {
	return s.arguments
}

func (s *ArgumentList) GetRanges() []protocol.Range {
	return s.ranges
}
