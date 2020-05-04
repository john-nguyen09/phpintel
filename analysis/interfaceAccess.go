package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// InterfaceAccess represents a reference to the part before ::
type InterfaceAccess struct {
	Expression
}

func newInterfaceAccess(document *Document, node *phrase.Phrase) *InterfaceAccess {
	interfaceAccess := &InterfaceAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
			Name:     document.getPhraseText(node),
		},
	}
	types := newTypeComposite()
	if node.Type == phrase.QualifiedName || node.Type == phrase.FullyQualifiedName {
		typeString := transformQualifiedName(node, document)
		typeString.SetFQN(document.currImportTable().GetClassReferenceFQN(typeString))
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
