package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

type PropertyAccess struct {
	Expression
}

func newPropertyAccess(document *Document, node *phrase.Phrase) HasTypes {
	propertyAccess := &PropertyAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok && p.Type == phrase.SimpleVariable {
		expression := scanForExpression(document, p)
		if variable, ok := expression.(*Variable); ok {
			propertyAccess.Scope = variable
		}
	}
	traverser.Advance()
	memberName := traverser.Advance()
	if p, ok := memberName.(*phrase.Phrase); ok && p.Type == phrase.MemberName {
		propertyAccess.Name = readMemberName(document, p)
	}
	return propertyAccess
}

func (s *PropertyAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *PropertyAccess) GetTypes() TypeComposite {
	return s.Type
}

func (s *PropertyAccess) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadPropertyAccess(serialiser *Serialiser) *PropertyAccess {
	return &PropertyAccess{
		Expression: ReadExpression(serialiser),
	}
}
