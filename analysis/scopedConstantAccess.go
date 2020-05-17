package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ScopedConstantAccess represents a reference to constant in class access, e.g. ::CONSTANT
type ScopedConstantAccess struct {
	Expression
}

var _ HasTypesHasScope = (*ScopedConstantAccess)(nil)

func newScopedConstantAccess(document *Document, node *phrase.Phrase) (HasTypes, bool) {
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
	return constantAccess, true
}

func (s *ScopedConstantAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ScopedConstantAccess) GetTypes() TypeComposite {
	// TODO: Look up constant types
	return s.Type
}

// MemberName returns the class constant name
func (s *ScopedConstantAccess) MemberName() string {
	return s.Name
}

// GetScopeTypes returns the types of the scope
func (s *ScopedConstantAccess) GetScopeTypes() TypeComposite {
	if s.Scope != nil {
		return s.Scope.GetTypes()
	}
	return newTypeComposite()
}
