package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Function contains information of functions
type Function struct {
	location protocol.Location

	Name   string `json:"Name"`
	Params []Parameter
}

func newFunction(document *Document, node *phrase.Phrase) Symbol {
	function := &Function{
		location: document.GetNodeLocation(node),
		Params:   make([]Parameter, 0),
	}
	document.pushVariableTable()

	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := util.IsOfPhraseTypes(child, []phrase.PhraseType{
			phrase.FunctionDeclarationHeader,
			phrase.MethodDeclarationHeader,
		}); ok {
			function.analyseHeader(document, p)
		}
		if p, ok := util.IsOfPhraseTypes(child, []phrase.PhraseType{
			phrase.FunctionDeclarationBody,
			phrase.MethodDeclarationBody,
		}); ok {
			scanForChildren(document, p)
		}
		child = traverser.Advance()
	}

	return function
}

func (s *Function) analyseHeader(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				{
					s.Name = document.GetTokenText(token)
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ParameterDeclarationList:
				{
					s.analyseParameterDeclarationList(document, p)
				}
			case phrase.Identifier:
				s.Name = document.GetPhraseText(p)
			}
		}
		child = traverser.Advance()
	}
}

func (s *Function) analyseParameterDeclarationList(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ParameterDeclaration {
			param := newParameter(document, p)
			s.Params = append(s.Params, *param)
		}

		child = traverser.Advance()
	}
}

func (s *Function) getLocation() protocol.Location {
	return s.location
}

func (s *Function) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	serialiser.WriteString(s.Name)
	serialiser.WriteInt(len(s.Params))
	for _, param := range s.Params {
		param.Write(serialiser)
	}
}

func ReadFunction(serialiser *Serialiser) *Function {
	function := Function{
		location: serialiser.ReadLocation(),
		Name:     serialiser.ReadString(),
		Params:   make([]Parameter, 0),
	}
	countParams := serialiser.ReadInt()
	for i := 0; i < countParams; i++ {
		function.Params = append(function.Params, ReadParameter(serialiser))
	}
	return &function
}
