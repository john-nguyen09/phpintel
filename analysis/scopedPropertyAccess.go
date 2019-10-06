package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// ScopedPropertyAccess represents a reference to property in
// scoped class access, e.g. ::$property
type ScopedPropertyAccess struct {
	Expression
}

func newScopedPropertyAccess(document *Document, parent symbolBlock, node *phrase.Phrase) Symbol {
	propertyAccess := &ScopedPropertyAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		classAccess := newClassAccess(document, parent, p)
		consumeIfIsConsumer(parent, classAccess)
		propertyAccess.Scope = &classAccess.Expression
	}
	traverser.Advance()
	thirdChild := traverser.Advance()
	if p, ok := thirdChild.(*phrase.Phrase); ok {
		propertyAccess.Name = analyseMemberName(document, p)
	}
	return propertyAccess
}

func (s *ScopedPropertyAccess) getLocation() lsp.Location {
	return s.Location
}
