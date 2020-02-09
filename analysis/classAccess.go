package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
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
	if node.Type == phrase.QualifiedName || node.Type == phrase.FullyQualifiedName {
		typeString := transformQualifiedName(node, document)
		typeString.SetFQN(document.GetImportTable().GetClassReferenceFQN(typeString))
		types.add(typeString)
	}
	if IsNameRelative(classAccess.Name) {
		relativeScope := newRelativeScope(document, classAccess.Location)
		types.merge(relativeScope.Types)
	}
	if IsNameParent(classAccess.Name) {
		parentScope := newParentScope(document, classAccess.Location)
		types.merge(parentScope.Types)
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

func (s *ClassAccess) GetName() string {
	return s.Name
}

func (s *ClassAccess) GetTypes() TypeComposite {
	return s.Type
}

func (s *ClassAccess) Serialise(e *storage.Encoder) {
	s.Expression.Serialise(e)
}

func ReadClassAccess(d *storage.Decoder) *ClassAccess {
	return &ClassAccess{
		Expression: ReadExpression(d),
	}
}
