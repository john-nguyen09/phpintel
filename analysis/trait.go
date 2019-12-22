package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Trait contains information of a trait
type Trait struct {
	location protocol.Location

	Name TypeString
}

func newTrait(document *Document, node *phrase.Phrase) Symbol {
	trait := &Trait{
		location: document.GetNodeLocation(node),
	}
	document.addClass(trait)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.TraitDeclarationHeader:
				document.addSymbol(trait)
				trait.analyseHeader(document, p)
			case phrase.TraitDeclarationBody:
				scanForChildren(document, p)
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func (s *Trait) analyseHeader(document *Document, traitHeader *phrase.Phrase) {
	traverser := util.NewTraverser(traitHeader)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			if token.Type == lexer.Name {
				s.Name = NewTypeString(document.GetTokenText(token))
				s.Name.SetNamespace(document.GetImportTable().namespace)
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

func (s *Trait) GetPrefixes() []string {
	scope, _ := GetScopeAndNameFromString(s.Name.GetFQN())
	prefixes := []string{""}
	if scope != "" {
		prefixes = append(prefixes, scope)
	}
	return prefixes
}

func (s *Trait) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	s.Name.Write(serialiser)
}

func ReadTrait(serialiser *Serialiser) *Trait {
	return &Trait{
		location: serialiser.ReadLocation(),
		Name:     ReadTypeString(serialiser),
	}
}
