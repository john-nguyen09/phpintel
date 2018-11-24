package entity

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type Class struct {
	Location   lsp.Range
	IsAbstract bool
	Name       string
	Extend     string
	Implements []string
	Traits     []string
}

func NewClass(node *phrase.Phrase, doc *PhpDoc) *Class {
	return &Class{
		Location:   util.NodeRange(node, doc.Text),
		IsAbstract: false,
		Name:       "",
		Extend:     "",
		Implements: nil,
		Traits:     nil}
}

func (c *Class) Consume(p *phrase.Phrase, doc *PhpDoc) {
	switch p.Type {
	case phrase.ClassDeclarationHeader:
		for _, child := range p.Children {
			if t, ok := child.(*lexer.Token); ok {
				if t.Type == lexer.Name {
					c.Name = string(util.GetTokenText(t, doc.Text))
				}

				if t.Type == lexer.Abstract {
					c.IsAbstract = true
				}
			}
		}
	case phrase.ClassBaseClause:
		for _, child := range p.Children {
			if childPhrase, ok := child.(*phrase.Phrase); ok && childPhrase.Type == phrase.QualifiedName {
				c.Extend = string(util.GetPhraseText(childPhrase, doc.Text))
				break
			}
		}
	case phrase.ClassInterfaceClause:
		for _, child := range p.Children {
			if childPhrase, ok := child.(*phrase.Phrase); ok && childPhrase.Type == phrase.QualifiedNameList {
				names := GetNames(childPhrase, doc)
				c.Implements = append(c.Implements, names...)
			}
		}
	case phrase.TraitUseClause:
		for _, child := range p.Children {
			if childPhrase, ok := child.(*phrase.Phrase); ok && childPhrase.Type == phrase.QualifiedNameList {
				names := GetNames(childPhrase, doc)

				c.Traits = append(c.Traits, names...)
			}
		}
	}
}
