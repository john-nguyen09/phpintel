package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Trait contains information of a trait
type Trait struct {
	location protocol.Location
	children []Symbol

	Name TypeString
}

var _ Symbol = (*Trait)(nil)
var _ BlockSymbol = (*Trait)(nil)

func newTrait(document *Document, node *ast.Node) Symbol {
	trait := &Trait{
		location: document.GetNodeLocation(node),
	}
	document.addClass(trait)
	document.addSymbol(trait)
	document.pushBlock(trait)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "name":
			trait.Name = NewTypeString(document.GetNodeText(child))
			trait.Name.SetNamespace(document.currImportTable().GetNamespace())
		case "declaration_list":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
	document.popBlock()
	return nil
}

func (s *Trait) GetLocation() protocol.Location {
	return s.location
}

func (s *Trait) GetName() string {
	return s.Name.original
}

func (s *Trait) GetDescription() string {
	// TODO: Docblock description
	return ""
}

func (s *Trait) GetCollection() string {
	return traitCollection
}

func (s *Trait) GetKey() string {
	return s.Name.fqn + KeySep + s.location.URI
}

func (s *Trait) GetIndexableName() string {
	return s.Name.GetOriginal()
}

func (s *Trait) GetIndexCollection() string {
	return traitCompletionIndex
}

func (s *Trait) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	s.Name.Write(e)
}

func ReadTrait(d *storage.Decoder) *Trait {
	return &Trait{
		location: d.ReadLocation(),
		Name:     ReadTypeString(d),
	}
}

func (s *Trait) addChild(child Symbol) {
	s.children = append(s.children, child)
}

func (s *Trait) GetChildren() []Symbol {
	return s.children
}
