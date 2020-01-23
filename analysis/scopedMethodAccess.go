package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ScopedMethodAccess represents a reference to method in class access, e.g. ::method()
type ScopedMethodAccess struct {
	Expression

	hasResolved bool
}

func newScopedMethodAccess(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	methodAccess := &ScopedMethodAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		classAccess := newClassAccess(document, p)
		document.addSymbol(classAccess)
		methodAccess.Scope = classAccess
	}
	document.addSymbol(methodAccess)
	traverser.Advance()
	thirdChild := traverser.Advance()
	methodAccess.Location = document.GetNodeLocation(thirdChild)
	if p, ok := thirdChild.(*phrase.Phrase); ok {
		methodAccess.Name = analyseMemberName(document, p)
	}
	child := traverser.Advance()
	var open *lexer.Token = nil
	var close *lexer.Token = nil
	hasArgs := false
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.ArgumentExpressionList {
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

func (s *ScopedMethodAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ScopedMethodAccess) Resolve(store *Store) {
	if s.hasResolved {
		return
	}
	s.hasResolved = true
	name := ""
	classScope := ""
	if hasName, ok := s.Scope.(HasName); ok {
		name = hasName.GetName()
	}
	if hasScope, ok := s.Scope.(HasScope); ok {
		classScope = hasScope.GetScope().GetFQN()
	}
	for _, scopeType := range s.ResolveAndGetScope(store).Resolve() {
		for _, class := range store.GetClasses(scopeType.GetFQN()) {
			for _, method := range GetClassMethods(store, class, s.Name,
				StaticMethodsScopeAware(NewSearchOptions(), classScope, name)) {
				s.Type.merge(method.GetReturnTypes())
			}
		}
	}
}

func (s *ScopedMethodAccess) GetTypes() TypeComposite {
	return s.Type
}

func (s *ScopedMethodAccess) ResolveToHasParams(store *Store, document *Document) []HasParams {
	hasParams := []HasParams{}
	for _, typeString := range s.ResolveAndGetScope(store).Resolve() {
		methods := store.GetMethods(typeString.GetFQN(), s.Name)
		for _, method := range methods {
			if !method.IsStatic {
				continue
			}
			hasParams = append(hasParams, method)
		}
	}
	return hasParams
}

func (s *ScopedMethodAccess) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadScopedMethodAccess(serialiser *Serialiser) *ScopedMethodAccess {
	return &ScopedMethodAccess{
		Expression: ReadExpression(serialiser),
	}
}
