package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func readMemberName(document *Document, traverser *util.Traverser) (string, protocol.Location) {
	next := traverser.Peek()
	location := protocol.Location{}
	for next != nil && (next.Type() == " " || next.Type() == "->") {
		if next.Type() == "->" {
			location.Range.Start = util.PointToPosition(next.EndPoint())
			location.Range.End = location.Range.Start
		}
		traverser.Advance()
		next = traverser.Peek()
	}
	memberName := traverser.Advance()
	name := ""
	if memberName.Type() == "name" {
		name = document.GetNodeText(memberName)
		location = document.GetNodeLocation(memberName)
	}
	return name, location
}
