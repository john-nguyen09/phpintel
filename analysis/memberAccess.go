package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func readMemberName(a analyser, document *Document, traverser *util.Traverser) (string, protocol.Location) {
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
			if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.SimpleVariable {
				name = document.getPhraseText(p)
				v, shouldAdd := newVariable(a, document, p, false)
				if shouldAdd {
					document.addSymbol(v)
				}
				return name, location
			}
		}
		location = document.GetNodeLocation(p)
	}
	return name, location
}

// MemberAccessExpression is the base struct for non-static member access
type MemberAccessExpression struct {
	Expression
}

// ScopeName returns the name of the accessed scope
func (m *MemberAccessExpression) ScopeName() string {
	var scopeName string
	if n, ok := m.Scope.(HasName); ok {
		scopeName = n.GetName()
	}
	return scopeName
}

// ScopeTypes returns the types of the scope, this should be resolved
// before calling this
func (m *MemberAccessExpression) ScopeTypes() TypeComposite {
	return m.Scope.GetTypes()
}
