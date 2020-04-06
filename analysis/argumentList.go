package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ArgumentList contains information of arguments in function-like call
type ArgumentList struct {
	location protocol.Location
	children []Symbol

	arguments []*ast.Node
	ranges    []protocol.Range
}

var _ BlockSymbol = (*ArgumentList)(nil)

func newArgumentList(document *Document, node *ast.Node) Symbol {
	argumentList := &ArgumentList{
		location: document.GetNodeLocation(node),
	}
	document.addSymbol(argumentList)
	document.pushBlock(argumentList)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	start := argumentList.location.Range.Start
	nodesToScan := []*ast.Node{}
	for child != nil {
		t := child.Type()
		if t == "," || (t == "ERROR" && document.GetNodeText(child) == ",") {
			end := util.PointToPosition(child.StartPoint())
			argumentList.ranges = append(argumentList.ranges, protocol.Range{
				Start: start,
				End:   end,
			})
			start = end
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
	document.popBlock()
	return argumentList
}

func (s *ArgumentList) GetLocation() protocol.Location {
	return s.location
}

// GetArguments returns the arguments
func (s *ArgumentList) GetArguments() []*ast.Node {
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
