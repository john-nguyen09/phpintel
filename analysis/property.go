package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Property contains information for properties
type Property struct {
	location           protocol.Location
	refLocation        protocol.Location
	description        string
	deprecatedTag      *tag
	Name               string
	Scope              TypeString
	VisibilityModifier VisibilityModifierValue
	isStatic           bool
	Types              TypeComposite
}

var _ HasScope = (*Property)(nil)
var _ Symbol = (*Property)(nil)
var _ SymbolReference = (*Property)(nil)
var _ MemberSymbol = (*Property)(nil)

func newPropertyFromPhpDocTag(document *Document, parent *Class, docTag tag, location protocol.Location) *Property {
	property := &Property{
		location:    location,
		refLocation: docTag.nameLocation,
		description: docTag.Description,

		Name:               docTag.Name,
		Scope:              parent.Name,
		VisibilityModifier: Public,
		isStatic:           false,
		Types:              typesFromPhpDoc(document, docTag.TypeString),
	}
	if property.Name == "" {
		return nil
	}
	return property
}

func newPropertyDeclaration(a analyser, document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	visibility := Public
	isStatic := false
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.MemberModifierList:
				visibility, isStatic, _ = getMemberModifier(p)
			case phrase.PropertyElementList:
				newPropertyList(a, document, p, visibility, isStatic)
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func newPropertyList(a analyser, document *Document, node *phrase.Phrase,
	visibility VisibilityModifierValue, isStatic bool) {
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.PropertyElement {
			property := newProperty(a, document, p, visibility, isStatic)
			document.addSymbol(property)
		}
	}
}

func newProperty(a analyser, document *Document, node *phrase.Phrase, visibility VisibilityModifierValue, isStatic bool) *Property {
	property := &Property{
		location:           document.GetNodeLocation(node),
		VisibilityModifier: visibility,
		isStatic:           isStatic,
	}
	parent := document.getLastClass()
	switch v := parent.(type) {
	case *Class:
		property.Scope = v.Name
	case *Trait:
		property.Scope = v.Name
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.PropertyInitialiser:
				scanForChildren(a, document, p)
			}
		} else if t, ok := child.(*lexer.Token); ok && t.Type == lexer.VariableName {
			property.Name = document.getTokenText(t)
			property.refLocation = document.GetNodeLocation(t)
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
	s.deprecatedTag = phpDoc.deprecated()
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
	if len(s.Name) == 0 {
		return ""
	}
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

func (s *Property) Serialise(e storage.Encoder) {
	e.WriteLocation(s.location)
	e.WriteString(s.Name)
	s.Scope.Write(e)
	e.WriteInt(int(s.VisibilityModifier))
	e.WriteBool(s.isStatic)
	s.Types.Write(e)
	e.WriteString(s.description)
	serialiseDeprecatedTag(e, s.deprecatedTag)
}

func ReadProperty(d storage.Decoder) *Property {
	return &Property{
		location:           d.ReadLocation(),
		Name:               d.ReadString(),
		Scope:              ReadTypeString(d),
		VisibilityModifier: VisibilityModifierValue(d.ReadInt()),
		isStatic:           d.ReadBool(),
		Types:              ReadTypeComposite(d),
		description:        d.ReadString(),
		deprecatedTag:      deserialiseDeprecatedTag(d),
	}
}

// ReferenceFQN returns the FQN of the property
func (s *Property) ReferenceFQN() string {
	return "." + s.Name
}

// ReferenceLocation returns the location of the property's name
func (s *Property) ReferenceLocation() protocol.Location {
	return s.refLocation
}

// IsStatic returns whether a property is static
func (s *Property) IsStatic() bool {
	return s.isStatic
}

// ScopeTypeString returns the class scope of the property
func (s *Property) ScopeTypeString() TypeString {
	return s.Scope
}

// Visibility returns the visibility modifier value for the property
func (s *Property) Visibility() VisibilityModifierValue {
	return s.VisibilityModifier
}
