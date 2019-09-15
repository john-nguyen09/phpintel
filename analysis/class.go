package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type Class struct {
	document *Document
	location lsp.Location

	Children   []Symbol
	Modifier   ClassModifier
	Name       TypeString
	Extends    TypeString
	Interfaces []TypeString
}

func ClassMemberDeclarationList(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	ScanForChildren(parent, node)

	return nil
}

func NewClass(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	class := &Class{
		document: document,
		location: document.GetNodeLocation(node),
		Children: []Symbol{},
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()

	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ClassDeclarationHeader:
				class.analyseHeader(p)
			case phrase.ClassDeclarationBody:
				ScanForChildren(class, p)
			}
		}

		child = traverser.Advance()
	}

	return class
}

func (s *Class) analyseHeader(classHeader *phrase.Phrase) {
	traverser := util.NewTraverser(classHeader)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				{
					s.Name = NewTypeString(util.GetNodeText(token, s.document.text))
				}
			case lexer.Abstract:
				{
					s.Modifier = Abstract
				}
			case lexer.Final:
				{
					s.Modifier = Final
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ClassBaseClause:
				{
					s.extends(p)
				}
			case phrase.ClassInterfaceClause:
				{
					s.implements(p)
				}
			}
		}

		child = traverser.Advance()
	}
}

func (s *Class) extends(p *phrase.Phrase) {
	traverser := util.NewTraverser(p)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName:
				{
					s.Extends = TransformQualifiedName(p, s.document)
				}
			}
		}

		child = traverser.Advance()
	}
}

func (s *Class) implements(p *phrase.Phrase) {
	traverser := util.NewTraverser(p)
	child := traverser.Peek()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedNameList {
			traverser.Descend()
			child = traverser.Advance()
			for child != nil {
				if p, ok = child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedName {
					s.Interfaces = append(s.Interfaces, TransformQualifiedName(p, s.document))
				}

				child = traverser.Advance()
			}

			break
		}

		traverser.Advance()
		child = traverser.Peek()
	}
}

func (s *Class) GetLocation() lsp.Location {
	return s.location
}

func (s *Class) GetDocument() *Document {
	return s.document
}

func (s *Class) Consume(other Symbol) {
	s.Children = append(s.Children, other)
}
