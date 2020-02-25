package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	sitter "github.com/smacker/go-tree-sitter"
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

// NameIndexable indicates a symbol is name indexable, i.e. have completion
type NameIndexable interface {
	GetIndexableName() string
	GetIndexCollection() string
}

type HasParams interface {
	GetParams() []*Parameter
	GetDescription() string
	GetNameLabel() string
}

type HasScope interface {
	GetScope() string
	IsScopeSymbol() bool
}

type serialisable interface {
	GetCollection() string
	GetKey() string
	Serialise(*storage.Encoder)
}

type HasParamsResolvable interface {
	ResolveToHasParams(ctx ResolveContext) []HasParams
}

func transformQualifiedName(n *sitter.Node, document *Document) TypeString {
	return NewTypeString(document.GetNodeText(n))
}
