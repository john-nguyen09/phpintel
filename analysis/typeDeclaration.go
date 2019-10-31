package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// TypeDeclaration is type declaration for a symbol
type TypeDeclaration struct {
	location protocol.Location

	Type TypeComposite
}

func newTypeDeclaration(document *Document, node *phrase.Phrase) *TypeDeclaration {
	typeDeclaration := &TypeDeclaration{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.QualifiedName {
				typeDeclaration.Type.add(transformQualifiedName(p, document))
			}
		}
		child = traverser.Advance()
	}

	return typeDeclaration
}
