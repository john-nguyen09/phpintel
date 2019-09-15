package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type Function struct {
	location lsp.Location
	document *Document

	Children []Symbol
	Name     string `json:"Name"`
	Params   []Parameter
}

func NewFunction(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	function := &Function{
		location: document.GetNodeLocation(node),
		document: document,
	}

	if len(node.Children) >= 1 {
		if p, ok := node.Children[0].(*phrase.Phrase); ok && p.Type == phrase.FunctionDeclarationHeader {
			function.analyseHeader(p)
		}
	}

	if len(node.Children) >= 2 {
		if p, ok := node.Children[1].(*phrase.Phrase); ok && p.Type == phrase.FunctionDeclarationBody {
			ScanForChildren(function, p)
		}
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
			param := NewParameter(s.document, s, p)
			s.Params = append(s.Params, *param)
		}

		child = traverser.Advance()
	}
}

func (s *Function) GetLocation() lsp.Location {
	return s.location
}

func (s *Function) GetDocument() *Document {
	return s.document
}

func (s *Function) Consume(other Symbol) {
	s.Children = append(s.Children, other)
}
