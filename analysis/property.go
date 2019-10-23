package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/indexer"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Property contains information for properties
type Property struct {
	location lsp.Location

	Name               string
	Scope              TypeString
	VisibilityModifier VisibilityModifierValue
	IsStatic           bool
}

func newPropertyDeclaration(document *Document, node *phrase.Phrase) Symbol {
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
				property := newProperty(document, p, visibility, isStatic)
				document.addSymbol(property)
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func newProperty(document *Document, node *phrase.Phrase, visibility VisibilityModifierValue, isStatic bool) *Property {
	property := &Property{
		location:           document.GetNodeLocation(node),
		VisibilityModifier: visibility,
		IsStatic:           isStatic,
	}
	parent := document.getLastClass()
	if theClass, ok := parent.(*Class); ok {
		property.Scope = theClass.Name
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.PropertyElement {
				property.Name = util.GetNodeText(p, document.GetText())
			}
		}
		child = traverser.Advance()
	}
	return property
}

func (s *Property) getLocation() lsp.Location {
	return s.location
}

func (s *Property) Serialise(serialiser *indexer.Serialiser) {
	util.WriteLocation(serialiser, s.location)
	serialiser.WriteString(s.Name)
	s.Scope.Write(serialiser)
	serialiser.WriteInt(int(s.VisibilityModifier))
	serialiser.WriteBool(s.IsStatic)
}

func ReadProperty(serialiser *indexer.Serialiser) *Property {
	return &Property{
		location:           util.ReadLocation(serialiser),
		Name:               serialiser.ReadString(),
		Scope:              ReadTypeString(serialiser),
		VisibilityModifier: VisibilityModifierValue(serialiser.ReadInt()),
		IsStatic:           serialiser.ReadBool(),
	}
}
