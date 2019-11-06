package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// ClassAccess represents a reference to the part before ::
type ClassAccess struct {
	Expression
}

func newClassAccess(document *Document, node *phrase.Phrase) *ClassAccess {
	classAccess := &ClassAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
			Name:     document.GetPhraseText(node),
		},
	}
	types := newTypeComposite()
	if node.Type == phrase.QualifiedName {
		types.add(transformQualifiedName(node, document))
	}
	classAccess.Type = types
	return classAccess
}

func analyseMemberName(document *Document, node *phrase.Phrase) string {
	if node.Type == phrase.ScopedMemberName {
		return document.GetPhraseText(node)
	}

	return ""
}

func (s *ClassAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ClassAccess) GetTypes() TypeComposite {
	return s.Type
}

func (s *ClassAccess) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadClassAccess(serialiser *Serialiser) *ClassAccess {
	return &ClassAccess{
		Expression: ReadExpression(serialiser),
	}
}
