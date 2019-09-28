package analysis

import (
	"github.com/sourcegraph/go-lsp"
)

// Expression represents a reference
type Expression struct {
	Type     SymbolType
	Scope    SymbolType
	Location lsp.Location
	Name     string
}
