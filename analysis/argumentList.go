package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	g "github.com/zyedidia/generic"
	"github.com/zyedidia/generic/hashset"
)

// ArgumentList contains information of arguments in function-like call
type ArgumentList struct {
	location protocol.Location
	children []Symbol

	arguments      []phrase.AstNode
	argumentRanges []protocol.Range
	ranges         []protocol.Range
}

var _ BlockSymbol = (*ArgumentList)(nil)
var IgnoreTokenSet *hashset.Set[lexer.TokenType]

func init() {
	ignoreTokens := []lexer.TokenType{
		lexer.Whitespace,
		lexer.OpenParenthesis,
		lexer.CloseParenthesis,
		lexer.Comment,
		lexer.DocumentCommentStart,
	}
	IgnoreTokenSet = hashset.New(uint64(len(ignoreTokens)), g.Equals[lexer.TokenType], func(t lexer.TokenType) uint64 {
		return g.HashUint8(uint8(t))
	})
	for _, t := range ignoreTokens {
		IgnoreTokenSet.Put(t)
	}
}

func newArgumentList(a analyser, document *Document, node *phrase.Phrase) Symbol {
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
				start = document.positionAt(token.Offset + token.Length)
				child = traverser.Advance()
				continue
			}
			if !IgnoreTokenSet.Has(token.Type) {
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
	for _, argument := range argumentList.GetArguments() {
		argumentList.argumentRanges = append(argumentList.argumentRanges, document.NodeRange(argument))
	}
	for _, n := range nodesToScan {
		scanNode(a, document, n)
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

// GetArgumentRanges returns the ranges of the arguments
// while this is not useful for providing signature help
// because the ranges ignore whitespaces, but this is
// useful for align signature annotations
func (s *ArgumentList) GetArgumentRanges() []protocol.Range {
	return s.argumentRanges
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
