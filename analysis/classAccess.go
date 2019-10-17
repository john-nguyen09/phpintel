package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// ClassAccess represents a reference to the part before ::
type ClassAccess struct {
	Expression
}

func newClassAccess(document *Document, node *phrase.Phrase) *ClassAccess {
	classAccess := &ClassAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
			Name:     util.GetNodeText(node, document.GetText()),
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
		return util.GetPhraseText(node, document.GetText())
	}

	return ""
}

func (s *ClassAccess) getLocation() lsp.Location {
	return s.Location
}

func (s *ClassAccess) getTypes() TypeComposite {
	return s.Type
}
