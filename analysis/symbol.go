package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	lsp "github.com/sourcegraph/go-lsp"
)

type VisibilityModifier int

const (
	Public    VisibilityModifier = 1 << iota
	Protected                    = 1 << iota
	Private                      = 1 << iota
)

type ClassModifier int

const (
	Abstract ClassModifier = iota
	Final                  = iota
)

type Symbol interface {
	GetLocation() lsp.Location
}

type HasConsume interface {
	Consume(symbol Symbol)
}

func TransformQualifiedName(p *phrase.Phrase, document *Document) TypeString {
	return NewTypeString(string(util.GetNodeText(p, document.GetText())))
}
