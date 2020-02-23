package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// ScopedConstantAccess represents a reference to constant in class access, e.g. ::CONSTANT
type ScopedConstantAccess struct {
	Expression
}

func newScopedConstantAccess(document *Document, node *sitter.Node) (HasTypes, bool) {
	constantAccess := &ScopedConstantAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	expr := scanForExpression(document, firstChild)
	document.addSymbol(expr)
	constantAccess.Scope = expr
	traverser.Advance()
	thirdChild := traverser.Advance()
	constantAccess.Location = document.GetNodeLocation(thirdChild)
	constantAccess.Name = analyseMemberName(document, thirdChild)
	return constantAccess, true
}

func (s *ScopedConstantAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ScopedConstantAccess) GetTypes() TypeComposite {
	// TODO: Look up constant types
	return s.Type
}

func (s *ScopedConstantAccess) Serialise(e *storage.Encoder) {
	s.Expression.Serialise(e)
}

func ReadScopedConstantAccess(d *storage.Decoder) *ScopedConstantAccess {
	return &ScopedConstantAccess{
		Expression: ReadExpression(d),
	}
}
