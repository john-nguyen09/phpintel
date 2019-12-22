package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type PropertyAccess struct {
	Expression

	hasResolved bool
}

func newPropertyAccess(document *Document, node *phrase.Phrase) (HasTypes, bool) {
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
	propertyAccess.Name, propertyAccess.Location = readMemberName(document, traverser)
	return propertyAccess, true
}

func (s *PropertyAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *PropertyAccess) Resolve(store *Store) {
	if s.hasResolved {
		return
	}
	for _, scopeType := range s.ResolveAndGetScope(store).Resolve() {
		for _, property := range store.GetProperties(scopeType.GetFQN(), "$"+s.Name) {
			s.Type.merge(property.Types)
		}
	}
	s.hasResolved = true
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
