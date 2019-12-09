package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
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

func (v VisibilityModifierValue) ToString() string {
	if v == Public {
		return "public"
	}
	if v == Private {
		return "private"
	}
	if v == Protected {
		return "protected"
	}
	return ""
}

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

// Symbol is a symbol
type Symbol interface {
	GetLocation() protocol.Location
}

// Definition symbol type is a type of symbol of (classes, interfaces,
// traits, properties, methods, class consts, defines, consts, functions)
type Definition interface {
	GetLocation() protocol.Location
	GetName() string
	GetDescription() string
}

// NameIndexable indicates a symbol is name indexable, i.e. have completion
type NameIndexable interface {
	GetIndexableName() string
	GetIndexCollection() string
	GetPrefix() string
}

type HasParams interface {
	GetParams() []*Parameter
	GetDescription() string
	GetNameLabel() string
}

type HasParamsResolvable interface {
	ResolveToHasParams(store *Store, document *Document) []HasParams
}

func transformQualifiedName(p *phrase.Phrase, document *Document) TypeString {
	return NewTypeString(string(document.GetNodeText(p)))
}
