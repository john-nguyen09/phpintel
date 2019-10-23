package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/indexer"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// ScopedPropertyAccess represents a reference to property in
// scoped class access, e.g. ::$property
type ScopedPropertyAccess struct {
	Expression
}

func newScopedPropertyAccess(document *Document, node *phrase.Phrase) hasTypes {
	propertyAccess := &ScopedPropertyAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	if p, ok := firstChild.(*phrase.Phrase); ok {
		classAccess := newClassAccess(document, p)
		document.addSymbol(classAccess)
		propertyAccess.Scope = classAccess
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

func (s *ScopedPropertyAccess) getTypes() TypeComposite {
	// TODO: Look up property types
	return s.Type
}

func (s *ScopedPropertyAccess) Serialiser(serialiser *indexer.Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadScopedPropertyAccess(serialiser *indexer.Serialiser) *ScopedPropertyAccess {
	return &ScopedPropertyAccess{
		Expression: ReadExpression(serialiser),
	}
}
