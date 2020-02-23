package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// ArgumentList contains information of arguments in function-like call
type ArgumentList struct {
	location protocol.Location

	arguments []*sitter.Node
	ranges    []protocol.Range
}

func newEmptyArgumentList(document *Document, open *sitter.Node, close *sitter.Node) *ArgumentList {
	closePos := util.PointToPosition(open.StartPoint())
	if close != nil {
		closePos = util.PointToPosition(close.StartPoint())
	}
	argumentList := &ArgumentList{
		location: protocol.Location{
			URI: document.GetURI(),
			Range: protocol.Range{
				Start: util.PointToPosition(open.StartPoint()),
				End:   closePos,
			},
		},
	}
	argumentList.ranges = append(argumentList.ranges, argumentList.location.Range)
	return argumentList
}

func newArgumentList(document *Document, node *sitter.Node) Symbol {
	argumentList := &ArgumentList{
		location: document.GetNodeLocation(node),
	}
	document.addSymbol(argumentList)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	start := argumentList.location.Range.Start
	for child != nil {
		if t := child.Type(); t == " " || t == "," {
			if t == "," {
				end := util.PointToPosition(child.StartPoint())
				argumentList.ranges = append(argumentList.ranges, protocol.Range{
					Start: start,
					End:   end,
				})
				start = end
			}
			child = traverser.Advance()
			continue
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
func (s *ArgumentList) GetArguments() []*sitter.Node {
	return s.arguments
}

func (s *ArgumentList) GetRanges() []protocol.Range {
	return s.ranges
}
