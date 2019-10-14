package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// ScopedMethodAccess represents a reference to method in class access, e.g. ::method()
type ScopedMethodAccess struct {
	Expression
}

func newScopedMethodAccess(document *Document, parent symbolBlock, node *phrase.Phrase) hasTypes {
	methodAccess := &ScopedMethodAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		classAccess := newClassAccess(document, parent, p)
		consumeIfIsConsumer(parent, classAccess)
		methodAccess.Scope = &classAccess.Expression
	}
	traverser.Advance()
	thirdChild := traverser.Advance()
	if p, ok := thirdChild.(*phrase.Phrase); ok {
		methodAccess.Name = analyseMemberName(document, p)
	}
	return methodAccess
}

func (s *ScopedMethodAccess) getLocation() lsp.Location {
	return s.Location
}

func (s *ScopedMethodAccess) getTypes() TypeComposite {
	// TODO: Look up method return type
	return s.Type
}
