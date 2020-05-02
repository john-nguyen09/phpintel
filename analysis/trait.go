package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
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

func newTrait(document *Document, node *phrase.Phrase) Symbol {
	trait := &Trait{
		location: document.GetNodeLocation(node),
	}
	document.addClass(trait)
	document.addSymbol(trait)
	document.pushBlock(trait)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.TraitDeclarationHeader:
				trait.analyseHeader(document, p)
			case phrase.TraitDeclarationBody:
				scanForChildren(document, p)
			}
		}
		child = traverser.Advance()
	}
	document.popBlock()
	return nil
}

func (s *Trait) analyseHeader(document *Document, traitHeader *phrase.Phrase) {
	traverser := util.NewTraverser(traitHeader)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			if token.Type == lexer.Name {
				s.Name = NewTypeString(document.getTokenText(token))
				s.Name.SetNamespace(document.currImportTable().GetNamespace())
			}
		}
		child = traverser.Advance()
	}
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
