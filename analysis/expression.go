package analysis

import (
	"github.com/sourcegraph/go-lsp"
)

// Expression represents a reference
type Expression struct {
	Type     TypeComposite
	Scope    *Expression
	Location lsp.Location
	Name     string
}

type hasTypes interface {
	getTypes() TypeComposite
}
