package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Parameter contains information of a function parameter
type Parameter struct {
	location    protocol.Location
	varLocation protocol.Location
	description string
	hasValue    bool

	Name  string        `json:"Name"`
	Type  TypeComposite `json:"Type"`
	Value string        `json:"Value"`
}

func newParameter(a analyser, document *Document, node *phrase.Phrase) *Parameter {
	param := &Parameter{
		location:    document.GetNodeLocation(node),
		varLocation: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	hasEqual := false
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.TypeDeclaration:
				typeDeclaration := newTypeDeclaration(document, p)
				for _, typeString := range typeDeclaration.Type.typeStrings {
					param.Type.add(typeString)
				}
				document.addSymbol(typeDeclaration)
			case phrase.ConstantAccessExpression:
				var (
					constAccess HasTypes
					shouldAdd   bool
				)
				if constAccess, shouldAdd = newConstantAccess(a, document, p); shouldAdd {
					document.addSymbol(constAccess)
				}
				if constAccess != nil && hasEqual {
					param.hasValue = true
					param.Value += document.getPhraseText(p)
				}
			default:
				if hasEqual {
					param.hasValue = true
					param.Value += document.getPhraseText(p)
				}
			}
		} else if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Equals:
				hasEqual = true
				traverser.SkipToken(lexer.Whitespace)
			case lexer.VariableName:
				param.Name = document.getTokenText(token)
				param.varLocation = document.GetNodeLocation(token)
			default:
				if hasEqual {
					param.hasValue = true
					param.Value += document.getTokenText(token)
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

func (s *Parameter) GetLocation() protocol.Location {
	return s.location
}

func (s Parameter) ToVariable() *Variable {
	return &Variable{
		Expression: Expression{
			Location: s.varLocation,
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
	e.WriteLocation(s.varLocation)
	e.WriteBool(s.hasValue)
	e.WriteString(s.Name)
	s.Type.Write(e)
	e.WriteString(s.Value)
}

func ReadParameter(d *storage.Decoder) *Parameter {
	return &Parameter{
		location:    d.ReadLocation(),
		varLocation: d.ReadLocation(),
		hasValue:    d.ReadBool(),
		Name:        d.ReadString(),
		Type:        ReadTypeComposite(d),
		Value:       d.ReadString(),
	}
}
