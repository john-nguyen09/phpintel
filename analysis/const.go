package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// Const contains information of constants
type Const struct {
	location protocol.Location

	Name  TypeString
	Value string
}

func newConstDeclaration(document *Document, node *sitter.Node) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if child.Type() == "const_element_list" {
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}

	return nil
}

func newConst(document *Document, node *sitter.Node) Symbol {
	constant := &Const{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	hasEquals := false
	for child != nil {
		switch child.Type() {
		case "name":
			constant.Name = NewTypeString(document.GetNodeText(child))
		case "=":
			hasEquals = true
		default:
			if hasEquals {
				constant.Value += document.GetNodeText(child)
			}
		}

		child = traverser.Advance()
	}
	constant.Name.SetNamespace(document.importTable.namespace)

	return constant
}

func (s *Const) GetLocation() protocol.Location {
	return s.location
}

func (s *Const) GetName() string {
	return s.Name.GetFQN()
}

func (s *Const) GetDescription() string {
	return s.GetName() + " = " + s.Value
}

func (s *Const) GetCollection() string {
	return constCollection
}

func (s *Const) GetKey() string {
	return s.GetName() + KeySep + s.location.URI
}

func (s *Const) GetIndexableName() string {
	return s.Name.GetOriginal()
}

func (s *Const) GetIndexCollection() string {
	return constCompletionIndex
}

func (s *Const) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	s.Name.Write(e)
	e.WriteString(s.Value)
}

func ReadConst(d *storage.Decoder) *Const {
	return &Const{
		location: d.ReadLocation(),
		Name:     ReadTypeString(d),
		Value:    d.ReadString(),
	}
}
