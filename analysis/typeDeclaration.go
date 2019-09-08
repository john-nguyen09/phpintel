package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type TypeDeclaration struct {
	location lsp.Location

	Type TypeComposite
}

func NewTypeDeclaration(document *Document, parent SymbolBlock, node *phrase.Phrase) *TypeDeclaration {
	typeDeclaration := &TypeDeclaration{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.QualifiedName {
				typeDeclaration.Type.Add(TransformQualifiedName(p, document))
			}
		}
		child = traverser.Advance()
	}

	return typeDeclaration
}

func (s *TypeDeclaration) GetLocation() lsp.Location {
	return s.location
}
