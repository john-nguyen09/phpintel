package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Class contains information of classes
type Class struct {
	document *Document
	location lsp.Location

	Modifier   ClassModifierValue
	Name       TypeString
	Extends    TypeString
	Interfaces []TypeString
}

func getMemberModifier(node *phrase.Phrase) (VisibilityModifierValue, bool, ClassModifierValue) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	visibilityModifier := Public
	classModifier := NoClassModifier
	isStatic := false
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Abstract:
				classModifier = Abstract
			case lexer.Final:
				classModifier = Final
			case lexer.Static:
				isStatic = true
			case lexer.Public:
				visibilityModifier = Public
			case lexer.Protected:
				visibilityModifier = Protected
			case lexer.Private:
				visibilityModifier = Private
			}
		}
		child = traverser.Advance()
	}

	return visibilityModifier, isStatic, classModifier
}

func newClass(document *Document, node *phrase.Phrase) Symbol {
	class := &Class{
		document: document,
		location: document.GetNodeLocation(node),
	}
	document.addClass(class)
	document.addSymbol(class)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()

	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ClassDeclarationHeader:
				class.analyseHeader(p)
			case phrase.ClassDeclarationBody:
				scanForChildren(document, p)
			}
		}

		child = traverser.Advance()
	}

	return nil
}

func (s *Class) analyseHeader(classHeader *phrase.Phrase) {
	traverser := util.NewTraverser(classHeader)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				{
					s.Name = newTypeString(util.GetNodeText(token, s.document.text))
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
					s.Extends = transformQualifiedName(p, s.document)
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
					s.Interfaces = append(s.Interfaces, transformQualifiedName(p, s.document))
				}

				child = traverser.Advance()
			}

			break
		}

		traverser.Advance()
		child = traverser.Peek()
	}
}

func (s *Class) getLocation() lsp.Location {
	return s.location
}

func (s *Class) GetCollection() string {
	return "class"
}

func (s *Class) GetKey() string {
	return s.Name.fqn + KeySep + s.document.GetURI()
}

func (s *Class) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	serialiser.WriteInt(int(s.Modifier))
	s.Name.Write(serialiser)
	s.Extends.Write(serialiser)
	serialiser.WriteInt(len(s.Interfaces))
	for _, theInterface := range s.Interfaces {
		theInterface.Write(serialiser)
	}
}

func ReadClass(serialiser *Serialiser) *Class {
	theClass := &Class{
		location: serialiser.ReadLocation(),
		Modifier: ClassModifierValue(serialiser.ReadInt()),
		Name:     ReadTypeString(serialiser),
		Extends:  ReadTypeString(serialiser),
	}
	numInterfaces := serialiser.ReadInt()
	for i := 0; i < numInterfaces; i++ {
		theClass.Interfaces = append(theClass.Interfaces, ReadTypeString(serialiser))
	}
	return theClass
}
