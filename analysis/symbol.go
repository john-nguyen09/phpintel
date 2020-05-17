package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
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

type BlockSymbol interface {
	Symbol
	GetChildren() []Symbol
	addChild(Symbol)
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

// SymbolReference is a reference to the symbol itself
type SymbolReference interface {
	Symbol
	ReferenceFQN() string
	ReferenceLocation() protocol.Location
}

type serialisable interface {
	GetCollection() string
	GetKey() string
	Serialise(*storage.Encoder)
}

type HasParamsResolvable interface {
	ResolveToHasParams(ctx ResolveContext) []HasParams
}

func transformQualifiedName(n *phrase.Phrase, document *Document) TypeString {
	return NewTypeString(document.GetNodeText(n))
}

type traverser struct {
	shouldStop  bool
	stopDescent bool
}

func newTraverser() traverser {
	return traverser{}
}

func (t *traverser) traverseDocument(document *Document, fn func(*traverser, Symbol)) {
	for _, child := range document.Children {
		t.traverseBlock(child, fn)
		if t.shouldStop {
			return
		}
	}
}

func (t *traverser) traverseBlock(s Symbol, fn func(*traverser, Symbol)) {
	fn(t, s)
	if t.shouldStop {
		return
	}
	if !t.stopDescent {
		if block, ok := s.(BlockSymbol); ok {
			for _, child := range block.GetChildren() {
				t.traverseBlock(child, fn)
			}
		}
	}
	t.stopDescent = false
}

func TraverseDocument(document *Document, preorder func(Symbol), postorder func(Symbol)) {
	for _, child := range document.Children {
		TraverseSymbol(child, preorder, postorder)
	}
}

func TraverseSymbol(s Symbol, preorder func(Symbol), postorder func(Symbol)) {
	preorder(s)
	if block, ok := s.(BlockSymbol); ok {
		for _, child := range block.GetChildren() {
			TraverseSymbol(child, preorder, postorder)
		}
	}
	if postorder != nil {
		postorder(s)
	}
}
