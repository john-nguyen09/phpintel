package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Class contains information of classes
type Class struct {
	description string
	children    []Symbol
	refLocation protocol.Location
	Location    protocol.Location

	Modifier   ClassModifierValue
	Name       TypeString
	Extends    TypeString
	Interfaces []TypeString
	Use        []TypeString

	deprecatedTag *tag
}

var _ HasScope = (*Class)(nil)
var _ Symbol = (*Class)(nil)
var _ BlockSymbol = (*Class)(nil)
var _ SymbolReference = (*Class)(nil)

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

func newClass(a analyser, document *Document, node *phrase.Phrase) Symbol {
	class := &Class{
		Location: document.GetNodeLocation(node),
	}
	document.addClass(class)
	phpDoc := document.getValidPhpDoc(class.Location)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	document.addSymbol(class)
	document.pushBlock(class)

	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ClassDeclarationHeader:
				class.analyseHeader(a, document, p, phpDoc)
			case phrase.ClassDeclarationBody:
				scanForChildren(a, document, p)
			}
		}
		child = traverser.Advance()
	}

	document.popBlock()

	return nil
}

func (s *Class) analyseHeader(a analyser, document *Document, classHeader *phrase.Phrase, phpDoc *phpDocComment) {
	traverser := util.NewTraverser(classHeader)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				s.Name = NewTypeString(document.getTokenText(token))
				s.Name.SetNamespace(document.currImportTable().GetNamespace())
				s.refLocation = document.GetNodeLocation(token)
				if phpDoc != nil {
					s.description = phpDoc.Description
					for _, propertyTag := range phpDoc.Properties {
						property := newPropertyFromPhpDocTag(document, s, propertyTag, phpDoc.GetLocation())
						document.addSymbol(property)
					}
					for _, propertyTag := range phpDoc.PropertyReads {
						property := newPropertyFromPhpDocTag(document, s, propertyTag, phpDoc.GetLocation())
						document.addSymbol(property)
					}
					for _, propertyTag := range phpDoc.PropertyWrites {
						property := newPropertyFromPhpDocTag(document, s, propertyTag, phpDoc.GetLocation())
						document.addSymbol(property)
					}
					for _, methodTag := range phpDoc.Methods {
						method := newMethodFromPhpDocTag(document, s, methodTag, phpDoc.GetLocation())
						document.addSymbol(method)
					}
					s.deprecatedTag = phpDoc.deprecated()
				}
			case lexer.Abstract:
				s.Modifier = Abstract
			case lexer.Final:
				s.Modifier = Final
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ClassBaseClause:
				s.extends(a, document, p)
			case phrase.ClassInterfaceClause:
				s.implements(document, p)
			}
		}
		child = traverser.Advance()
	}
}

func (s *Class) extends(a analyser, document *Document, p *phrase.Phrase) {
	traverser := util.NewTraverser(p)
	child := traverser.Advance()
	var classAccessNode *phrase.Phrase = nil
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName, phrase.FullyQualifiedName:
				{
					s.Extends = transformQualifiedName(p, document)
					s.Extends.SetFQN(document.currImportTable().GetClassReferenceFQN(s.Extends))
					classAccessNode = p
				}
			}
		}

		child = traverser.Advance()
	}

	if classAccessNode != nil {
		classAccess := newClassAccess(a, document, classAccessNode)
		document.addSymbol(classAccess)
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
				if p, ok = child.(*phrase.Phrase); ok && (p.Type == phrase.QualifiedName || p.Type == phrase.FullyQualifiedName) {
					typeString := transformQualifiedName(p, document)
					typeString.SetFQN(document.currImportTable().GetClassReferenceFQN(typeString))
					s.Interfaces = append(s.Interfaces, typeString)

					interfaceAccess := newInterfaceAccess(document, p)
					document.addSymbol(interfaceAccess)
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
	return s.description
}

func (s *Class) GetCollection() string {
	return classCollection
}

func (s *Class) GetKey() string {
	return s.Name.GetFQN() + KeySep + s.Location.URI
}

func (s *Class) GetIndexableName() string {
	return s.Name.GetOriginal()
}

func (s *Class) GetIndexCollection() string {
	return classCompletionIndex
}

func (s *Class) GetScope() string {
	return s.Name.GetNamespace()
}

func (s *Class) IsScopeSymbol() bool {
	return false
}

func (s *Class) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.Location)
	e.WriteLocation(s.refLocation)
	e.WriteInt(int(s.Modifier))
	e.WriteString(s.description)
	s.Name.Write(e)
	s.Extends.Write(e)
	e.WriteInt(len(s.Interfaces))
	for _, theInterface := range s.Interfaces {
		theInterface.Write(e)
	}
	e.WriteInt(len(s.Use))
	for _, use := range s.Use {
		use.Write(e)
	}
	serialiseDeprecatedTag(e, s.deprecatedTag)
}

func ReadClass(d *storage.Decoder) *Class {
	theClass := &Class{
		Location:    d.ReadLocation(),
		refLocation: d.ReadLocation(),
		Modifier:    ClassModifierValue(d.ReadInt()),
		description: d.ReadString(),
		Name:        ReadTypeString(d),
		Extends:     ReadTypeString(d),
	}
	numInterfaces := d.ReadInt()
	for i := 0; i < numInterfaces; i++ {
		theClass.Interfaces = append(theClass.Interfaces, ReadTypeString(d))
	}
	numUse := d.ReadInt()
	for i := 0; i < numUse; i++ {
		theClass.Use = append(theClass.Use, ReadTypeString(d))
	}
	theClass.deprecatedTag = deserialiseDeprecatedTag(d)
	return theClass
}

func (s *Class) GetConstructor(store *Store) *Method {
	methods := GetClassMethods(store, s, "__construct", NewSearchOptions())
	if len(methods) > 0 {
		return methods[0]
	}
	return nil
}

func (s *Class) AddUse(name TypeString) {
	s.Use = append(s.Use, name)
}

func (s *Class) addChild(child Symbol) {
	s.children = append(s.children, child)
}

func (s *Class) GetChildren() []Symbol {
	return s.children
}

// ReferenceFQN returns the FQN of the class' name
func (s *Class) ReferenceFQN() string {
	return s.Name.GetFQN()
}

// ReferenceLocation returns the location of the class' name
func (s *Class) ReferenceLocation() protocol.Location {
	return s.refLocation
}

func (s *Class) findProp(name string) *Property {
	var (
		prop *Property
		ok   bool
	)
	for _, child := range s.children {
		if prop, ok = child.(*Property); ok && prop.Name == name {
			break
		}
		prop = nil
	}
	return prop
}
