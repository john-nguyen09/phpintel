package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// Trait contains information of a trait
type Trait struct {
	location protocol.Location

	Name TypeString
}

func newTrait(document *Document, node *sitter.Node) Symbol {
	trait := &Trait{
		location: document.GetNodeLocation(node),
	}
	document.addClass(trait)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "name":
			document.addSymbol(trait)
			trait.Name = NewTypeString(document.GetNodeText(child))
			trait.Name.SetNamespace(document.currImportTable().GetNamespace())
		case "declaration_list":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
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
