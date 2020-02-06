package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
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

func newParameter(document *Document, node *phrase.Phrase) *Parameter {
	param := &Parameter{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	hasEqual := false
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.TypeDeclaration:
				{
					typeDeclaration := newTypeDeclaration(document, p)
					for _, typeString := range typeDeclaration.Type.typeStrings {
						param.Type.add(typeString)
					}
					document.addSymbol(typeDeclaration)
				}
			case phrase.ConstantAccessExpression:
				if constAccess, shouldAdd := newConstantAccess(document, p); shouldAdd {
					document.addSymbol(constAccess)
				}
			}

			if hasEqual {
				param.hasValue = true
				param.Value += document.GetPhraseText(p)
			}
		} else if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Equals:
				{
					hasEqual = true
				}
			case lexer.VariableName:
				{
					param.Name = document.GetTokenText(token)
				}
			default:
				if hasEqual {
					param.hasValue = true
					param.Value += document.GetTokenText(token)
				}
			}
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

func (s *Parameter) Write(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	serialiser.WriteBool(s.hasValue)
	serialiser.WriteString(s.Name)
	s.Type.Write(serialiser)
	serialiser.WriteString(s.Value)
}

func ReadParameter(serialiser *Serialiser) *Parameter {
	return &Parameter{
		location: serialiser.ReadLocation(),
		hasValue: serialiser.ReadBool(),
		Name:     serialiser.ReadString(),
		Type:     ReadTypeComposite(serialiser),
		Value:    serialiser.ReadString(),
	}
}
