package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// Property contains information for properties
type Property struct {
	location    protocol.Location
	description string

	Name               string
	Scope              TypeString
	VisibilityModifier VisibilityModifierValue
	IsStatic           bool
	Types              TypeComposite
}

var _ HasScope = (*Property)(nil)
var _ Symbol = (*Property)(nil)

func newPropertyFromPhpDocTag(document *Document, parent *Class, docTag tag, location protocol.Location) *Property {
	property := &Property{
		location:    location,
		description: docTag.Description,

		Name:               docTag.Name,
		Scope:              parent.Name,
		VisibilityModifier: Public,
		IsStatic:           false,
		Types:              typesFromPhpDoc(document, docTag.TypeString),
	}
	return property
}

func newPropertyDeclaration(document *Document, node *sitter.Node) Symbol {
	traverser := util.NewTraverser(node)
	visibility := Public
	isStatic := false
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "visibility_modifier":
			visibility = getMemberModifier(child)
		case "static_modifier":
			isStatic = true
		case "property_element":
			property := newProperty(document, child, visibility, isStatic)
			document.addSymbol(property)
		}
		child = traverser.Advance()
	}
	return nil
}

func newProperty(document *Document, node *sitter.Node, visibility VisibilityModifierValue, isStatic bool) *Property {
	property := &Property{
		location:           document.GetNodeLocation(node),
		VisibilityModifier: visibility,
		IsStatic:           isStatic,
	}
	parent := document.getLastClass()
	switch v := parent.(type) {
	case *Class:
		property.Scope = v.Name
	case *Trait:
		property.Scope = v.Name
	}
	if theClass, ok := parent.(*Class); ok {
		property.Scope = theClass.Name
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if child.Type() == "variable_name" {
			property.Name = document.GetNodeText(child)
		}
		child = traverser.Advance()
	}
	phpDoc := document.getValidPhpDoc(property.location)
	if phpDoc != nil {
		property.applyPhpDoc(document, *phpDoc)
	}
	return property
}

func (s *Property) GetLocation() protocol.Location {
	return s.location
}

func (s *Property) GetName() string {
	return s.Name
}

func (s *Property) GetDescription() string {
	return s.description
}

func (s *Property) applyPhpDoc(document *Document, phpDoc phpDocComment) {
	tags := phpDoc.Vars
	for _, tag := range tags {
		s.Types.merge(typesFromPhpDoc(document, tag.TypeString))
		s.description = tag.Description
		break
	}
}

func (s *Property) GetCollection() string {
	return propertyCollection
}

func (s *Property) GetKey() string {
	return s.Scope.GetFQN() + KeySep + s.Name + KeySep + s.location.URI + s.location.Range.String()
}

func (s *Property) GetIndexableName() string {
	return string([]rune(s.Name)[1:])
}

func (s *Property) GetIndexCollection() string {
	return propertyCompletionIndex
}

func (s *Property) GetScope() string {
	return s.Scope.GetFQN()
}

func (s *Property) IsScopeSymbol() bool {
	return true
}

func (s *Property) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	e.WriteString(s.Name)
	s.Scope.Write(e)
	e.WriteInt(int(s.VisibilityModifier))
	e.WriteBool(s.IsStatic)
	s.Types.Write(e)
}

func ReadProperty(d *storage.Decoder) *Property {
	return &Property{
		location:           d.ReadLocation(),
		Name:               d.ReadString(),
		Scope:              ReadTypeString(d),
		VisibilityModifier: VisibilityModifierValue(d.ReadInt()),
		IsStatic:           d.ReadBool(),
		Types:              ReadTypeComposite(d),
	}
}
