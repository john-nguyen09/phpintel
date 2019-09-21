package analysis

import (
	"github.com/sourcegraph/go-lsp"
)

// Reference represents a reference
type Reference struct {
	Type     SymbolType
	Scope    SymbolType
	Location lsp.Location
	Name     string
}

type hasReferences interface {
	GetReferences() []Reference
}
