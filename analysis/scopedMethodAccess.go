package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// ScopedMethodAccess represents a reference to method in class access, e.g. ::method()
type ScopedMethodAccess struct {
	Expression
}

func newScopedMethodAccess(document *Document, node *phrase.Phrase) hasTypes {
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
	traverser.Advance()
	thirdChild := traverser.Advance()
	methodAccess.Location = document.GetNodeLocation(thirdChild)
	if p, ok := thirdChild.(*phrase.Phrase); ok {
		methodAccess.Name = analyseMemberName(document, p)
	}
	return methodAccess
}

func (s *ScopedMethodAccess) getLocation() protocol.Location {
	return s.Location
}

func (s *ScopedMethodAccess) getTypes() TypeComposite {
	// TODO: Look up method return type
	return s.Type
}

func (s *ScopedMethodAccess) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadScopedMethodAccess(serialiser *Serialiser) *ScopedMethodAccess {
	return &ScopedMethodAccess{
		Expression: ReadExpression(serialiser),
	}
}
