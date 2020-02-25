package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	sitter "github.com/smacker/go-tree-sitter"
)

// InterfaceAccess represents a reference to the part before ::
type InterfaceAccess struct {
	Expression
}

func newInterfaceAccess(document *Document, node *sitter.Node) *InterfaceAccess {
	interfaceAccess := &InterfaceAccess{
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
	interfaceAccess.Type = types
	return interfaceAccess
}

func (s *InterfaceAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *InterfaceAccess) GetTypes() TypeComposite {
	return s.Type
}
