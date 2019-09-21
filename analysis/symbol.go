package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	lsp "github.com/sourcegraph/go-lsp"
)

type VisibilityModifierValue int

const (
	Public    VisibilityModifierValue = iota
	Protected                         = iota
	Private                           = iota
)

type ClassModifierValue int

const (
	NoClassModifier ClassModifierValue = iota
	Abstract        ClassModifierValue = iota
	Final                              = iota
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
