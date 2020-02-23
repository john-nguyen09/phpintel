package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	sitter "github.com/smacker/go-tree-sitter"
)

// ConstantAccess represents a reference to constant access
type ConstantAccess struct {
	Expression
}

func newDirectoryConstantAccess(document *Document, token *sitter.Node) Symbol {
	constantAccess := &ConstantAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(token),
		},
	}
	constantAccess.readName(document, token)
	return constantAccess
}
func newConstantAccess(document *Document, node *sitter.Node) (HasTypes, bool) {
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

func (s *ConstantAccess) Serialise(e *storage.Encoder) {
	s.Expression.Serialise(e)
}

func ReadConstantAccess(d *storage.Decoder) *ConstantAccess {
	return &ConstantAccess{
		Expression: ReadExpression(d),
	}
}
