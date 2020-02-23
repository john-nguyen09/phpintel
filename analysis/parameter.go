package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// Parameter contains information of a function parameter
type Parameter struct {
	location    protocol.Location
	description string
	hasValue    bool

	Name  string        `json:"Name"`
	Type  TypeComposite `json:"Type"`
	Value string        `json:"Value"`
}

func newParameter(document *Document, node *sitter.Node) *Parameter {
	param := &Parameter{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	hasEqual := false
	for child != nil {
		switch child.Type() {
		case "type_name":
			{
				typeDeclaration := newTypeDeclaration(document, child)
				for _, typeString := range typeDeclaration.Type.typeStrings {
					param.Type.add(typeString)
				}
				document.addSymbol(typeDeclaration)
			}
		case "variable_name":
			param.Name = document.GetNodeText(child)
		case "=":
			hasEqual = true
		default:
			if hasEqual {
				param.hasValue = true
				param.Value += document.GetNodeText(child)
			}
		}

		if hasEqual {
			param.hasValue = true
			param.Value += document.GetNodeText(child)
		}
		child = traverser.Advance()
	}

	return param
}

func (s *Parameter) GetDescription() string {
	return s.description
}

func (s Parameter) ToVariable() *Variable {
	return &Variable{
		Expression: Expression{
			Location: s.location,
			Type:     s.Type,
			Name:     s.Name,
			Scope:    nil,
		},
		description:        s.description,
		canReferenceGlobal: false,
	}
}

func (s Parameter) HasValue() bool {
	return s.hasValue
}

func (s *Parameter) Write(e *storage.Encoder) {
	e.WriteLocation(s.location)
	e.WriteBool(s.hasValue)
	e.WriteString(s.Name)
	s.Type.Write(e)
	e.WriteString(s.Value)
}

func ReadParameter(d *storage.Decoder) *Parameter {
	return &Parameter{
		location: d.ReadLocation(),
		hasValue: d.ReadBool(),
		Name:     d.ReadString(),
		Type:     ReadTypeComposite(d),
		Value:    d.ReadString(),
	}
}
