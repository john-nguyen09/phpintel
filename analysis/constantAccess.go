package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	sitter "github.com/smacker/go-tree-sitter"
)

// ConstantAccess represents a reference to constant access
type ConstantAccess struct {
	Expression
}

func processQualifiedName(document *Document, node *sitter.Node) (HasTypes, bool) {
	prev := node.PrevSibling()
	if prev != nil && prev.Type() == "instanceof" {
		return newTypeDeclaration(document, node), true
	}
	next := node.NextSibling()
	if next != nil && next.Type() == "::" {
		c := newClassAccess(document, node)
		return c, true
	}
	constantAccess := &ConstantAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	constantAccess.readName(document, node)
	return constantAccess, true
}

func (s *ConstantAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ConstantAccess) readName(document *Document, node *sitter.Node) {
	s.Name = document.GetNodeText(node)
}

func (s *ConstantAccess) GetTypes() TypeComposite {
	// TODO: look up constant type
	return s.Type
}
