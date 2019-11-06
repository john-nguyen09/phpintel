package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

type MethodAccess struct {
	Expression
}

func newMethodAccess(document *Document, node *phrase.Phrase) HasTypes {
	methodAccess := &MethodAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok && p.Type == phrase.SimpleVariable {
		expression := scanForExpression(document, p)
		if variable, ok := expression.(*Variable); ok {
			methodAccess.Scope = variable
		}
	}
	traverser.Advance()
	memberName := traverser.Advance()
	if p, ok := memberName.(*phrase.Phrase); ok && p.Type == phrase.MemberName {
		methodAccess.Name = readMemberName(document, p)
	}
	return methodAccess
}

func (s *MethodAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *MethodAccess) GetTypes() TypeComposite {
	// TODO: Lookup method return types
	return s.Type
}

func (s *MethodAccess) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadMethodAccess(serialiser *Serialiser) *MethodAccess {
	return &MethodAccess{
		Expression: ReadExpression(serialiser),
	}
}
