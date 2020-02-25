package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// ClassConst contains information of class constants
type ClassConst struct {
	location protocol.Location

	Name  string
	Value string
	Scope TypeString
}

var _ HasScope = (*ClassConst)(nil)
var _ Symbol = (*ClassConst)(nil)

func newClassConst(document *Document, node *sitter.Node) Symbol {
	classConst := &ClassConst{
		location: document.GetNodeLocation(node),
	}

	parent := document.getLastClass()
	if theClass, ok := parent.(*Class); ok {
		classConst.Scope = theClass.Name
	}

	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	hasEquals := false
	for child != nil {
		switch child.Type() {
		case "=":
			hasEquals = true
			traverser.SkipToken(" ")
		default:
			if hasEquals {
				classConst.Value += document.GetNodeText(child)
			} else {
				if child.Type() == "name" {
					classConst.Name = document.GetNodeText(child)
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
	// TODO: Implement docblock description
	return ""
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
	e.WriteString(s.Name)
	e.WriteString(s.Value)
	s.Scope.Write(e)
}

func ReadClassConst(d *storage.Decoder) *ClassConst {
	return &ClassConst{
		location: d.ReadLocation(),
		Name:     d.ReadString(),
		Value:    d.ReadString(),
		Scope:    ReadTypeString(d),
	}
}
