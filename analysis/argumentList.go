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
	children []Symbol

	arguments []phrase.AstNode
	ranges    []protocol.Range
}

var _ BlockSymbol = (*ArgumentList)(nil)

func newEmptyArgumentList(document *Document, open *lexer.Token, close *lexer.Token) *ArgumentList {
	closePos := document.positionAt(open.Offset)
	if close != nil {
		closePos = document.positionAt(close.Offset + close.Length)
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
	document.pushBlock(argumentList)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	start := argumentList.location.Range.Start
	nodesToScan := []*phrase.Phrase{}
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			if token.Type == lexer.Comma {
				end := document.positionAt(token.Offset)
				argumentList.ranges = append(argumentList.ranges, protocol.Range{
					Start: start,
					End:   end,
				})
				start = end
				child = traverser.Advance()
				continue
			}
			if token.Type != lexer.Whitespace {
				argumentList.arguments = append(argumentList.arguments, token)
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			nodesToScan = append(nodesToScan, p)
			argumentList.arguments = append(argumentList.arguments, p)
		}
		child = traverser.Advance()
	}
	argumentList.ranges = append(argumentList.ranges, protocol.Range{
		Start: start,
		End:   argumentList.location.Range.End,
	})
	for _, n := range nodesToScan {
		scanNode(document, n)
	}
	document.popBlock()
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

func (s *ArgumentList) addChild(child Symbol) {
	s.children = append(s.children, child)
}

func (s *ArgumentList) GetChildren() []Symbol {
	return s.children
}
