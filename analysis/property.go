package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Property contains information for properties
type Property struct {
	location protocol.Location

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
				if len(p.Children) > 0 {
					firstChild := p.Children[0]
					if t, ok := firstChild.(*lexer.Token); ok && t.Type == lexer.VariableName {
						property.Name = document.GetTokenText(t)
					}
				}
			}
		}
		child = traverser.Advance()
	}
	return property
}

func (s *Property) getLocation() protocol.Location {
	return s.location
}

func (s *Property) GetCollection() string {
	return "property"
}

func (s *Property) GetKey() string {
	return s.Scope.fqn + KeySep + s.Name
}

func (s *Property) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	serialiser.WriteString(s.Name)
	s.Scope.Write(serialiser)
	serialiser.WriteInt(int(s.VisibilityModifier))
	serialiser.WriteBool(s.IsStatic)
}

func ReadProperty(serialiser *Serialiser) *Property {
	return &Property{
		location:           serialiser.ReadLocation(),
		Name:               serialiser.ReadString(),
		Scope:              ReadTypeString(serialiser),
		VisibilityModifier: VisibilityModifierValue(serialiser.ReadInt()),
		IsStatic:           serialiser.ReadBool(),
	}
}
