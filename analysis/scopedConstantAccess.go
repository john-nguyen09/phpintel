package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// ScopedConstantAccess represents a reference to constant in class access, e.g. ::CONSTANT
type ScopedConstantAccess struct {
	Expression
}

func newScopedConstantAccess(document *Document, node *phrase.Phrase) HasTypes {
	constantAccess := &ScopedConstantAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		classAccess := newClassAccess(document, p)
		document.addSymbol(classAccess)
		constantAccess.Scope = classAccess
	}
	traverser.Advance()
	thirdChild := traverser.Advance()
	constantAccess.Location = document.GetNodeLocation(thirdChild)
	if p, ok := thirdChild.(*phrase.Phrase); ok {
		constantAccess.Name = analyseMemberName(document, p)
	}
	return constantAccess
}

func (s *ScopedConstantAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ScopedConstantAccess) GetTypes() TypeComposite {
	// TODO: Look up constant types
	return s.Type
}

func (s *ScopedConstantAccess) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadScopedConstantAccess(serialiser *Serialiser) *ScopedConstantAccess {
	return &ScopedConstantAccess{
		Expression: ReadExpression(serialiser),
	}
}
