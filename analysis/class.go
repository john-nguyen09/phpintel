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
	children []Symbol

	Modifier   ClassModifier
	Name       TypeString
	Extends    TypeString
	Interfaces []TypeString
}

func NewClass(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	class := &Class{
		document: document,
		location: document.GetNodeLocation(node),
	}

	if classHeader, ok := node.Children[0].(*phrase.Phrase); ok && classHeader.Type == phrase.ClassDeclarationHeader {
		class.analyseHeader(classHeader)
	}
	if len(node.Children) >= 2 {
		if classBody, ok := node.Children[1].(*phrase.Phrase); ok {
			ScanForChildren(class, classBody)
		}
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

func (s *Class) GetChildren() []Symbol {
	return s.children
}

func (s *Class) Consume(other Symbol) {
	s.children = append(s.children, other)
}
