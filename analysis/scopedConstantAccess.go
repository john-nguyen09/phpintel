package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ScopedConstantAccess represents a reference to constant in class access, e.g. ::CONSTANT
type ScopedConstantAccess struct {
	Expression
}

func newScopedConstantAccess(document *Document, node *ast.Node) (HasTypes, bool) {
	constantAccess := &ScopedConstantAccess{
		Expression: Expression{},
	}
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	classAccess := newClassAccess(document, firstChild)
	document.addSymbol(classAccess)
	constantAccess.Scope = classAccess
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
