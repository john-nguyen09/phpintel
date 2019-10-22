package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Trait contains information of a trait
type Trait struct {
	document *Document
	location lsp.Location

	Name TypeString
}

func newTrait(document *Document, node *phrase.Phrase) Symbol {
	trait := &Trait{
		document: document,
		location: document.GetNodeLocation(node),
	}
	document.addClass(trait)
	if traitHeader, ok := node.Children[0].(*phrase.Phrase); ok && traitHeader.Type == phrase.TraitDeclarationHeader {
		trait.analyseHeader(traitHeader)
	}

	if len(node.Children) >= 2 {
		if classBody, ok := node.Children[1].(*phrase.Phrase); ok {
			scanForChildren(document, classBody)
		}
	}

	return trait
}

func (s *Trait) analyseHeader(traitHeader *phrase.Phrase) {
	traverser := util.NewTraverser(traitHeader)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			if token.Type == lexer.Name {
				s.Name = newTypeString(util.GetNodeText(token, s.document.GetText()))
			}
		}

		child = traverser.Advance()
	}
}

func (s *Trait) getDocument() *Document {
	return s.document
}

func (s *Trait) getLocation() lsp.Location {
	return s.location
}
