package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// Interface contains information of interfaces
type Interface struct {
	location    protocol.Location
	description string

	Name    TypeString
	Extends []TypeString
}

var _ HasScope = (*Interface)(nil)
var _ Symbol = (*Interface)(nil)

func newInterface(document *Document, node *sitter.Node) Symbol {
	theInterface := &Interface{
		location: document.GetNodeLocation(node),
	}
	document.addClass(theInterface)
	document.addSymbol(theInterface)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "name":
			theInterface.Name = NewTypeString(document.GetNodeText(child))
		case "interface_base_clause":
			theInterface.extends(document, child)
		case "body":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
	theInterface.Name.SetNamespace(document.importTable.namespace)
	return nil
}

func (s *Interface) extends(document *Document, node *sitter.Node) {
	traverser := util.NewTraverser(node)
	child := traverser.Peek()
	for child != nil {
		if child.Type() == "qualified_name" {
			s.Extends = append(s.Extends, transformQualifiedName(child, document))
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
