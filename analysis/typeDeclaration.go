package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// TypeDeclaration is type declaration for a symbol
type TypeDeclaration struct {
	Expression
}

func newTypeDeclaration(document *Document, node *sitter.Node) *TypeDeclaration {
	typeDeclaration := &TypeDeclaration{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	if node.Type() == "type" {
		typeString := transformQualifiedName(node, document)
		typeDeclaration.Name = typeString.GetOriginal()
		typeString.SetFQN(document.currImportTable().GetClassReferenceFQN(typeString))
		typeDeclaration.Type.add(typeString)
		return typeDeclaration
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "name":
			typeString := transformQualifiedName(child, document)
			typeDeclaration.Name = typeString.GetOriginal()
			typeString.SetFQN(document.currImportTable().GetClassReferenceFQN(typeString))
			typeDeclaration.Type.add(typeString)
		}
		child = traverser.Advance()
	}

	return typeDeclaration
}

func (s *TypeDeclaration) GetLocation() protocol.Location {
	return s.Location
}

func (s *TypeDeclaration) GetTypes() TypeComposite {
	return s.Type
}
