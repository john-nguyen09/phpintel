package analysis

import (
	"github.com/sourcegraph/go-lsp"
)

type Reference struct {
	Type     SymbolType
	Scope    SymbolType
	Location lsp.Location
	Name     string
}

type HasReferences interface {
	GetReferences() []Reference
}
