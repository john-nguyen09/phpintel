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
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "qualified_name":
			typeString := transformQualifiedName(child, document)
			typeDeclaration.Name = typeString.GetOriginal()
			typeString.SetFQN(document.GetImportTable().GetClassReferenceFQN(typeString))
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
