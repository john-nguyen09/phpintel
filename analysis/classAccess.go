package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	sitter "github.com/smacker/go-tree-sitter"
)

// ClassAccess represents a reference to the part before ::
type ClassAccess struct {
	Expression
}

var _ (HasTypes) = (*ClassAccess)(nil)

func newClassAccess(document *Document, node *sitter.Node) *ClassAccess {
	classAccess := &ClassAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
			Name:     document.GetNodeText(node),
		},
	}
	types := newTypeComposite()
	if node.Type() == "qualified_name" {
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

func analyseMemberName(document *Document, node *sitter.Node) string {
	switch node.Type() {
	case "name", "variable_name":
		return document.GetNodeText(node)
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
