package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	lsp "github.com/sourcegraph/go-lsp"
)

// VisibilityModifierValue is a value of visibility modifier (public, protected, private)
type VisibilityModifierValue int

const (
	// Public indicates public visibility
	Public VisibilityModifierValue = iota
	// Protected indicates protected visibility
	Protected = iota
	// Private indicates private visibility
	Private = iota
)

// ClassModifierValue is a value of class modifier (abstract, final)
type ClassModifierValue int

const (
	// NoClassModifier indicates there is no class modifier
	NoClassModifier ClassModifierValue = iota
	// Abstract indicates there is an abstract keyword
	Abstract = iota
	// Final indicates there is a final keyword
	Final = iota
)

// KeySep is the separator when constructing symbol key
const KeySep = "\x00"

// Symbol is a symbol
type Symbol interface {
	getLocation() lsp.Location
}

func transformQualifiedName(p *phrase.Phrase, document *Document) TypeString {
	return newTypeString(string(util.GetNodeText(p, document.GetText())))
}
