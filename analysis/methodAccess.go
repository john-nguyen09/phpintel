package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type MethodAccess struct {
	Expression

	hasResolved bool
}

func newMethodAccess(document *Document, node *phrase.Phrase) HasTypes {
	methodAccess := &MethodAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		expression := scanForExpression(document, p)
		if expression != nil {
			methodAccess.Scope = expression
		}
	}
	traverser.Advance()
	methodAccess.Name, methodAccess.Location = readMemberName(document, traverser)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ArgumentExpressionList:
				newArgumentList(document, p)
			}
		}
		child = traverser.Advance()
	}
	return methodAccess
}

func (s *MethodAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *MethodAccess) Resolve(store *Store) {
	if s.hasResolved {
		return
	}
	for _, scopeType := range s.ResolveAndGetScope(store).Resolve() {
		for _, method := range store.GetMethods(scopeType.GetFQN(), s.Name) {
			s.Type.merge(method.returnTypes)
		}
	}
	s.hasResolved = true
}

func (s *MethodAccess) GetTypes() TypeComposite {
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
