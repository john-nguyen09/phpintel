package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Function contains information of functions
type Function struct {
	location lsp.Location
	document *Document

	Children []Symbol
	Name     string `json:"Name"`
	Params   []Parameter
}

func newFunction(document *Document, parent symbolBlock, node *phrase.Phrase) Symbol {
	function := &Function{
		location: document.GetNodeLocation(node),
		document: document,

		Children: make([]Symbol, 0),
		Params:   make([]Parameter, 0),
	}

	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := util.IsOfPhraseTypes(child, []phrase.PhraseType{
			phrase.FunctionDeclarationHeader,
			phrase.MethodDeclarationHeader,
		}); ok {
			function.analyseHeader(p)
		}
		if p, ok := util.IsOfPhraseTypes(child, []phrase.PhraseType{
			phrase.FunctionDeclarationBody,
			phrase.MethodDeclarationBody,
		}); ok {
			scanForChildren(function, p)
		}
		child = traverser.Advance()
	}

	return function
}

func (s *Function) analyseHeader(node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				{
					s.Name = util.GetNodeText(token, s.document.text)
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ParameterDeclarationList:
				{
					s.analyseParameterDeclarationList(p)
				}
			case phrase.Identifier:
				s.Name = util.GetNodeText(p, s.document.text)
			}
		}
		child = traverser.Advance()
	}
}

func (s *Function) analyseParameterDeclarationList(node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ParameterDeclaration {
			param := newParameter(s.document, s, p)
			s.Params = append(s.Params, *param)
		}

		child = traverser.Advance()
	}
}

func (s *Function) getLocation() lsp.Location {
	return s.location
}

func (s *Function) getDocument() *Document {
	return s.document
}

func (s *Function) consume(other Symbol) {
	s.Children = append(s.Children, other)
}