package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func readMemberName(document *Document, traverser *util.Traverser) (string, protocol.Location) {
	next := traverser.Peek()
	location := protocol.Location{}
	for nextToken, ok := next.(*lexer.Token); ok && (nextToken.Type == lexer.Whitespace || nextToken.Type == lexer.Arrow); {
		if nextToken.Type == lexer.Arrow {
			location.Range.Start = document.positionAt(nextToken.Offset + nextToken.Length)
			location.Range.End = location.Range.Start
		}
		traverser.Advance()
		next = traverser.Peek()
		nextToken, ok = next.(*lexer.Token)
	}
	memberName := traverser.Advance()
	name := ""
	if p, ok := memberName.(*phrase.Phrase); ok && p.Type == phrase.MemberName {
		for _, child := range p.Children {
			if t, ok := child.(*lexer.Token); ok && t.Type == lexer.Name {
				name = document.getTokenText(t)
			}
		}
		location = document.GetNodeLocation(p)
	}
	return name, location
}
