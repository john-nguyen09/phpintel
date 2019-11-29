package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func readMemberName(document *Document, traverser util.Traverser) (string, protocol.Location) {
	next := traverser.Peek()
	var startRange *protocol.Range = nil
	for nextToken, ok := next.(*lexer.Token); ok && nextToken.Type == lexer.Whitespace; {
		if startRange == nil {
			theRange := document.nodeRange(next)
			startRange = &theRange
		}
		traverser.Advance()
		next = traverser.Peek()
		nextToken, ok = next.(*lexer.Token)
	}
	memberName := traverser.Advance()
	location := protocol.Location{}
	name := ""
	if p, ok := memberName.(*phrase.Phrase); ok && p.Type == phrase.MemberName {
		for _, child := range p.Children {
			if t, ok := child.(*lexer.Token); ok && t.Type == lexer.Name {
				name = document.GetTokenText(t)
			}
		}
		location = document.GetNodeLocation(p)
		if startRange != nil {
			location.Range.Start = startRange.Start
		}
	}
	return name, location
}
