package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// ScopedConstantAccess represents a reference to constant in class access, e.g. ::CONSTANT
type ScopedConstantAccess struct {
	Expression
}

func newScopedConstantAccess(document *Document, parent symbolBlock, node *phrase.Phrase) hasTypes {
	constantAccess := &ScopedConstantAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		classAccess := newClassAccess(document, parent, p)
		consumeIfIsConsumer(parent, classAccess)
		constantAccess.Scope = &classAccess.Expression
	}
	traverser.Advance()
	thirdChild := traverser.Advance()
	if p, ok := thirdChild.(*phrase.Phrase); ok {
		constantAccess.Name = analyseMemberName(document, p)
	}
	return constantAccess
}

func (s *ScopedConstantAccess) getLocation() lsp.Location {
	return s.Location
}

func (s *ScopedConstantAccess) getTypes() TypeComposite {
	// TODO: Look up constant types
	return s.Type
}
