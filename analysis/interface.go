package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Interface contains information of interfaces
type Interface struct {
	location      protocol.Location
	refLocation   protocol.Location
	children      []Symbol
	description   string
	deprecatedTag *tag
	Name          TypeString
	Extends       []TypeString
}

var _ HasScope = (*Interface)(nil)
var _ Symbol = (*Interface)(nil)
var _ BlockSymbol = (*Interface)(nil)
var _ SymbolReference = (*Interface)(nil)

func newInterface(a analyser, document *Document, node *phrase.Phrase) Symbol {
	theInterface := &Interface{
		location: document.GetNodeLocation(node),
	}
	document.addClass(theInterface)
	phpDoc := document.getValidPhpDoc(theInterface.location)
	document.addSymbol(theInterface)
	document.pushBlock(theInterface)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.InterfaceDeclarationHeader:
				theInterface.analyseHeader(document, p, phpDoc)
			case phrase.InterfaceDeclarationBody:
				scanForChildren(a, document, p)
			}
		}
		child = traverser.Advance()
	}
	theInterface.Name.SetNamespace(document.currImportTable().GetNamespace())
	document.popBlock()
	return nil
}

func (s *Interface) analyseHeader(document *Document, node *phrase.Phrase, phpDoc *phpDocComment) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				s.Name = NewTypeString(document.getTokenText(token))
				s.refLocation = document.GetNodeLocation(token)
				if phpDoc != nil {
					s.description = phpDoc.Description
					s.deprecatedTag = phpDoc.deprecated()
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.InterfaceBaseClause:
				s.extends(document, p)
			}
		}
		child = traverser.Advance()
	}
}

func (s *Interface) extends(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Peek()
	for child != nil {
		if _, ok := child.(*phrase.Phrase); ok {
			traverser, err := traverser.Descend()
			if err != nil {
				continue
			}
			child = traverser.Advance()
			for child != nil {
				if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedName {
					s.Extends = append(s.Extends, transformQualifiedName(p, document))
				}
				child = traverser.Advance()
			}
		}
		traverser.Advance()
		child = traverser.Peek()
	}
}

func (s *Interface) GetLocation() protocol.Location {
	return s.location
}

func (s *Interface) GetDescription() string {
	return s.description
}

func (s *Interface) GetCollection() string {
	return interfaceCollection
}

func (s *Interface) GetKey() string {
	return s.Name.GetFQN() + KeySep + s.location.URI
}

func (s *Interface) GetIndexableName() string {
	return s.Name.GetOriginal()
}

func (s *Interface) GetIndexCollection() string {
	return interfaceCompletionIndex
}

func (s *Interface) GetScope() string {
	return s.Name.GetNamespace()
}

func (s *Interface) IsScopeSymbol() bool {
	return false
}

func (s *Interface) Serialise(e storage.Encoder) {
	e.WriteLocation(s.location)
	s.Name.Write(e)
	e.WriteInt(len(s.Extends))
	for _, extend := range s.Extends {
		extend.Write(e)
	}
	e.WriteString(s.description)
	serialiseDeprecatedTag(e, s.deprecatedTag)
}

func ReadInterface(d storage.Decoder) *Interface {
	theInterface := &Interface{
		location: d.ReadLocation(),
		Name:     ReadTypeString(d),
	}
	countExtends := d.ReadInt()
	for i := 0; i < countExtends; i++ {
		theInterface.Extends = append(theInterface.Extends, ReadTypeString(d))
	}
	theInterface.description = d.ReadString()
	theInterface.deprecatedTag = deserialiseDeprecatedTag(d)
	return theInterface
}

func (s *Interface) addChild(child Symbol) {
	s.children = append(s.children, child)
}

func (s *Interface) GetChildren() []Symbol {
	return s.children
}

// ReferenceFQN returns the interface's FQN for reference index
func (s *Interface) ReferenceFQN() string {
	return s.Name.GetFQN()
}

// ReferenceLocation returns the location of the interface's name
func (s *Interface) ReferenceLocation() protocol.Location {
	return s.refLocation
}

func (s *Interface) findProp(name string) *Property {
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
