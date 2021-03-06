package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ClassConst contains information of class constants
type ClassConst struct {
	refLocation protocol.Location
	location    protocol.Location
	description string

	Name               string
	Value              string
	Scope              TypeString
	VisibilityModifier VisibilityModifierValue

	deprecatedTag *tag
}

var _ HasScope = (*ClassConst)(nil)
var _ Symbol = (*ClassConst)(nil)
var _ SymbolReference = (*ClassConst)(nil)
var _ MemberSymbol = (*ClassConst)(nil)

func newClassConstDeclaration(a analyser, document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	visibility := Public
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.MemberModifierList:
				visibility, _, _ = getMemberModifier(p)
			case phrase.ClassConstElementList:
				newClassConstList(a, document, p, visibility)
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func newClassConstList(a analyser, document *Document, node *phrase.Phrase, visibility VisibilityModifierValue) {
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ClassConstElement {
			property := newClassConst(a, document, p, visibility)
			document.addSymbol(property)
		}
	}
}

func newClassConst(a analyser, document *Document, node *phrase.Phrase, visibility VisibilityModifierValue) Symbol {
	classConst := &ClassConst{
		location:           document.GetNodeLocation(node),
		VisibilityModifier: visibility,
	}
	phpDoc := document.getValidPhpDoc(classConst.location)
	lastClass := document.getLastClass()
	if theClass, ok := lastClass.(*Class); ok {
		classConst.Scope = theClass.Name
		classConst.Scope.SetNamespace(document.currImportTable().GetNamespace())
	} else if theInterface, ok := lastClass.(*Interface); ok {
		classConst.Scope = theInterface.Name
		classConst.Scope.SetNamespace(document.currImportTable().GetNamespace())
	} else if trait, ok := lastClass.(*Trait); ok {
		classConst.Scope = trait.Name
		classConst.Scope.SetNamespace(document.currImportTable().GetNamespace())
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	hasEquals := false
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Equals:
				hasEquals = true
				traverser.SkipToken(lexer.Whitespace)
			default:
				if hasEquals {
					classConst.Value += document.getTokenText(token)
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			if hasEquals {
				classConst.Value += document.getPhraseText(p)
				scanNode(a, document, p)
			} else {
				switch p.Type {
				case phrase.Identifier:
					classConst.Name = document.getPhraseText(p)
					classConst.refLocation = document.GetNodeLocation(p)
					if phpDoc != nil {
						classConst.description = phpDoc.Description
						classConst.deprecatedTag = phpDoc.deprecated()
					}
				}
			}
		}
		child = traverser.Advance()
	}
	return classConst
}

func (s *ClassConst) GetLocation() protocol.Location {
	return s.location
}

func (s *ClassConst) GetName() string {
	return s.Name
}

func (s *ClassConst) GetDescription() string {
	return s.description
}

func (s *ClassConst) GetCollection() string {
	return classConstCollection
}

func (s *ClassConst) GetKey() string {
	return s.Scope.fqn + KeySep + s.Name
}

func (s *ClassConst) GetIndexableName() string {
	return s.GetName()
}

func (s *ClassConst) GetIndexCollection() string {
	return classConstCompletionIndex
}

func (s *ClassConst) GetScope() string {
	return s.Scope.GetFQN()
}

func (s *ClassConst) IsScopeSymbol() bool {
	return true
}

func (s *ClassConst) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	e.WriteString(s.description)
	e.WriteString(s.Name)
	e.WriteString(s.Value)
	s.Scope.Write(e)
	serialiseDeprecatedTag(e, s.deprecatedTag)
}

func ReadClassConst(d *storage.Decoder) *ClassConst {
	return &ClassConst{
		location:      d.ReadLocation(),
		description:   d.ReadString(),
		Name:          d.ReadString(),
		Value:         d.ReadString(),
		Scope:         ReadTypeString(d),
		deprecatedTag: deserialiseDeprecatedTag(d),
	}
}

func (s *ClassConst) ReferenceFQN() string {
	return "." + s.Name
}

func (s *ClassConst) ReferenceLocation() protocol.Location {
	return s.refLocation
}

// IsStatic is always static for class const
func (s *ClassConst) IsStatic() bool {
	return true
}

// ScopeTypeString returns the class const's class name
func (s *ClassConst) ScopeTypeString() TypeString {
	return s.Scope
}

// Visibility returns the visibility modifier of the class const
func (s *ClassConst) Visibility() VisibilityModifierValue {
	return s.VisibilityModifier
}
