package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type MethodAccess struct {
	Expression

	hasResolved bool
}

func newMethodAccess(document *Document, node *phrase.Phrase) (HasTypes, bool) {
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
	document.addSymbol(methodAccess)
	child := traverser.Advance()
	var open *lexer.Token = nil
	var close *lexer.Token = nil
	hasArgs := false
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ArgumentExpressionList:
				newArgumentList(document, p)
				hasArgs = true
				break
			}
		} else if t, ok := child.(*lexer.Token); ok {
			switch t.Type {
			case lexer.OpenParenthesis:
				open = t
			case lexer.CloseParenthesis:
				close = t
			}
		}
		child = traverser.Advance()
	}
	if !hasArgs {
		args := newEmptyArgumentList(document, open, close)
		document.addSymbol(args)
	}
	return methodAccess, false
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

func (s *MethodAccess) ResolveToHasParams(store *Store, document *Document) []HasParams {
	hasParams := []HasParams{}
	for _, typeString := range s.ResolveAndGetScope(store).Resolve() {
		methods := store.GetMethods(typeString.GetFQN(), s.Name)
		for _, method := range methods {
			if method.IsStatic {
				continue
			}
			hasParams = append(hasParams, method)
		}
	}
	return hasParams
}

func (s *MethodAccess) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadMethodAccess(serialiser *Serialiser) *MethodAccess {
	return &MethodAccess{
		Expression: ReadExpression(serialiser),
	}
}
