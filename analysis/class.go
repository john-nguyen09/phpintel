package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Class contains information of classes
type Class struct {
	Location protocol.Location

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
		Location: document.GetNodeLocation(node),
	}
	document.addClass(class)
	document.addSymbol(class)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()

	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ClassDeclarationHeader:
				class.analyseHeader(document, p)
			case phrase.ClassDeclarationBody:
				scanForChildren(document, p)
			}
		}

		child = traverser.Advance()
	}

	return nil
}

func (s *Class) analyseHeader(document *Document, classHeader *phrase.Phrase) {
	traverser := util.NewTraverser(classHeader)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				{
					s.Name = NewTypeString(document.GetTokenText(token))
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
					s.extends(document, p)
				}
			case phrase.ClassInterfaceClause:
				{
					s.implements(document, p)
				}
			}
		}

		child = traverser.Advance()
	}

	s.Name.SetNamespace(document.importTable.namespace)
}

func (s *Class) extends(document *Document, p *phrase.Phrase) {
	traverser := util.NewTraverser(p)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName:
				{
					s.Extends = transformQualifiedName(p, document)
				}
			}
		}

		child = traverser.Advance()
	}
}

func (s *Class) implements(document *Document, p *phrase.Phrase) {
	traverser := util.NewTraverser(p)
	child := traverser.Peek()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedNameList {
			traverser, _ = traverser.Descend()
			child = traverser.Advance()
			for child != nil {
				if p, ok = child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedName {
					s.Interfaces = append(s.Interfaces, transformQualifiedName(p, document))
				}

				child = traverser.Advance()
			}

			break
		}

		traverser.Advance()
		child = traverser.Peek()
	}
}

func (s *Class) GetLocation() protocol.Location {
	return s.Location
}

func (s *Class) GetName() string {
	return s.Name.original
}

func (s *Class) GetDescription() string {
	// TODO: implement php docblock
	return ""
}

func (s *Class) GetCollection() string {
	return classCollection
}

func (s *Class) GetKey() string {
	return s.Name.fqn + KeySep + s.Location.URI
}

func (s *Class) GetIndexableName() string {
	return s.Name.GetOriginal()
}

func (s *Class) GetIndexCollection() string {
	return classCompletionIndex
}

func (s *Class) GetPrefix() string {
	return ""
}

func (s *Class) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.Location)
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
		Location: serialiser.ReadLocation(),
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
