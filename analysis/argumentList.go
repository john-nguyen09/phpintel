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

func newArgumentList(document *Document, node *sitter.Node) Symbol {
	argumentList := &ArgumentList{
		location: document.GetNodeLocation(node),
	}
	document.addSymbol(argumentList)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	start := argumentList.location.Range.Start
	nodesToScan := []*sitter.Node{}
	for child != nil {
		t := child.Type()
		if t == " " || t == "," {
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
		if t != "(" && t != ")" {
			argumentList.arguments = append(argumentList.arguments, child)
			switch t {
			default:
				nodesToScan = append(nodesToScan, child)
			}
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
