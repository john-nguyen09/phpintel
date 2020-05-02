package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// ClassAccess represents a reference to the part before ::
type ClassAccess struct {
	Expression

	isResolved bool
}

var _ (HasTypes) = (*ClassAccess)(nil)

func newClassAccess(document *Document, node *phrase.Phrase) *ClassAccess {
	classAccess := &ClassAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
			Name:     document.getPhraseText(node),
		},
	}
	types := newTypeComposite()
	switch node.Type {
	case phrase.QualifiedName, phrase.FullyQualifiedName:
		typeString := transformQualifiedName(node, document)
		typeString.SetFQN(document.currImportTable().GetClassReferenceFQN(typeString))
		types.add(typeString)
	case phrase.SimpleVariable:
		expr := scanForExpression(document, node)
		if expr != nil {
			classAccess.Scope = expr
		}
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
		return document.getPhraseText(node)
	}

	return ""
}

func (s *ClassAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ClassAccess) GetName() string {
	return s.Name
}

func (s *ClassAccess) Resolve(ctx ResolveContext) {
	if s.isResolved {
		return
	}
	s.isResolved = true
	s.Type.merge(s.ResolveAndGetScope(ctx))
}

func (s *ClassAccess) GetTypes() TypeComposite {
	return s.Type
}
