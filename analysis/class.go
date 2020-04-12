package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
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
}

var _ HasScope = (*Class)(nil)
var _ Symbol = (*Class)(nil)
var _ BlockSymbol = (*Class)(nil)
var _ SymbolReference = (*Class)(nil)

func getMemberModifier(node *sitter.Node) VisibilityModifierValue {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	visibilityModifier := Public
	for child != nil {
		switch child.Type() {
		case "public":
			visibilityModifier = Public
		case "protected":
			visibilityModifier = Protected
		case "private":
			visibilityModifier = Private
		}
		child = traverser.Advance()
	}

	return visibilityModifier
}

func getClassModifier(node *sitter.Node) ClassModifierValue {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	c := NoClassModifier
	for child != nil {
		switch child.Type() {
		case "final":
			c = Final
		case "abstract":
			c = Abstract
		}
		child = traverser.Advance()
	}
	return c
}

func newClass(document *Document, node *sitter.Node) Symbol {
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
		switch child.Type() {
		case "name":
			class.Name = NewTypeString(document.GetNodeText(child))
			class.Name.SetNamespace(document.currImportTable().GetNamespace())
			class.refLocation = document.GetNodeLocation(child)
			if phpDoc != nil {
				class.description = phpDoc.Description
				for _, propertyTag := range phpDoc.Properties {
					property := newPropertyFromPhpDocTag(document, class, propertyTag, phpDoc.GetLocation())
					document.addSymbol(property)
				}
				for _, propertyTag := range phpDoc.PropertyReads {
					property := newPropertyFromPhpDocTag(document, class, propertyTag, phpDoc.GetLocation())
					document.addSymbol(property)
				}
				for _, propertyTag := range phpDoc.PropertyWrites {
					property := newPropertyFromPhpDocTag(document, class, propertyTag, phpDoc.GetLocation())
					document.addSymbol(property)
				}
				for _, methodTag := range phpDoc.Methods {
					method := newMethodFromPhpDocTag(document, class, methodTag, phpDoc.GetLocation())
					document.addSymbol(method)
				}
			}
		case "class_modifier":
			class.analyseClassModifier(document, child)
		case "class_base_clause":
			class.extends(document, child)
		case "class_interface_clause":
			class.implements(document, child)

		case "declaration_list":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
	document.popBlock()

	return nil
}

func (s *Class) analyseClassModifier(document *Document, n *sitter.Node) {
	s.Modifier = getClassModifier(n)
}

func (s *Class) extends(document *Document, p *sitter.Node) {
	traverser := util.NewTraverser(p)
	child := traverser.Advance()
	var classAccessNode *sitter.Node = nil
	for child != nil {
		switch child.Type() {
		case "qualified_name":
			{
				s.Extends = transformQualifiedName(child, document)
				s.Extends.SetFQN(document.currImportTable().GetClassReferenceFQN(s.Extends))
				classAccessNode = child
			}
		}

		child = traverser.Advance()
	}

	if classAccessNode != nil {
		classAccess := newClassAccess(document, classAccessNode)
		document.addSymbol(classAccess)
	}
}

func (s *Class) implements(document *Document, p *sitter.Node) {
	traverser := util.NewTraverser(p)
	child := traverser.Advance()
	for child != nil {
		if child.Type() == "qualified_name" {
			typeString := transformQualifiedName(child, document)
			typeString.SetFQN(document.currImportTable().GetClassReferenceFQN(typeString))
			s.Interfaces = append(s.Interfaces, typeString)

			interfaceAccess := newInterfaceAccess(document, child)
			document.addSymbol(interfaceAccess)
		}
		child = traverser.Advance()
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
