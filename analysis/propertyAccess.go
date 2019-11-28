package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type PropertyAccess struct {
	Expression
}

func newPropertyAccess(document *Document, node *phrase.Phrase) HasTypes {
	propertyAccess := &PropertyAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		expression := scanForExpression(document, p)
		if expression != nil {
			propertyAccess.Scope = expression
		}
	}
	traverser.Advance()

	propertyAccess.Name, propertyAccess.Location = readMemberName(document, traverser)
	return propertyAccess
}

func (s *PropertyAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *PropertyAccess) Resolve(store *Store) {

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
