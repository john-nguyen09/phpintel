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
	location    protocol.Location
	children    []Symbol
	description string

	Name    TypeString
	Extends []TypeString
}

var _ HasScope = (*Interface)(nil)
var _ Symbol = (*Interface)(nil)
var _ BlockSymbol = (*Interface)(nil)

func newInterface(document *Document, node *phrase.Phrase) Symbol {
	theInterface := &Interface{
		location: document.GetNodeLocation(node),
	}
	document.addClass(theInterface)
	document.addSymbol(theInterface)
	document.pushBlock(theInterface)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.InterfaceBaseClause:
				theInterface.extends(document, p)
			case phrase.InterfaceDeclarationBody:
				scanForChildren(document, p)
			}
		} else if t, ok := child.(*lexer.Token); ok {
			switch t.Type {
			case lexer.Name:
				theInterface.Name = NewTypeString(document.GetNodeText(t))
			}
		}
		child = traverser.Advance()
	}
	theInterface.Name.SetNamespace(document.currImportTable().GetNamespace())
	document.popBlock()
	return nil
}

func (s *Interface) extends(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Peek()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			traverser, err := traverser.Descend()
			if err != nil {
				continue
			}
			child = traverser.Advance()
			for child != nil {
				if p, ok = child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedName {
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

func (s *Interface) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	s.Name.Write(e)
	e.WriteInt(len(s.Extends))
	for _, extend := range s.Extends {
		extend.Write(e)
	}
}

func ReadInterface(d *storage.Decoder) *Interface {
	theInterface := &Interface{
		location: d.ReadLocation(),
		Name:     ReadTypeString(d),
	}
	countExtends := d.ReadInt()
	for i := 0; i < countExtends; i++ {
		theInterface.Extends = append(theInterface.Extends, ReadTypeString(d))
	}
	return theInterface
}

func (s *Interface) addChild(child Symbol) {
	s.children = append(s.children, child)
}

func (s *Interface) GetChildren() []Symbol {
	return s.children
}
