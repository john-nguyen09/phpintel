package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// TypeDeclaration is type declaration for a symbol
type TypeDeclaration struct {
	Expression
}

func newTypeDeclaration(document *Document, node *phrase.Phrase) *TypeDeclaration {
	typeDeclaration := &TypeDeclaration{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName, phrase.FullyQualifiedName:
				typeString := transformQualifiedName(p, document)
				typeDeclaration.Name = typeString.GetOriginal()
				typeString.SetFQN(document.GetImportTable().GetClassReferenceFQN(typeString))
				typeDeclaration.Type.add(typeString)
			}
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
