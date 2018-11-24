package entity

import (
	"github.com/sourcegraph/go-lsp"
)

type Function struct {
	Location lsp.Range
	Name     string
	Type     string
	Params   []Param
}
