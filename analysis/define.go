package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type Define struct {
	document *Document
	location lsp.Location
	children []Symbol

	Name  string
	Value string
}

func NewDefine(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	define := &Define{
		document: document,
		location: document.GetNodeLocation(node),
	}

	ScanForChildren(define, node)

	return define
}

func (s *Define) GetLocation() lsp.Location {
	return s.location
}

func (s *Define) GetDocument() *Document {
	return s.document
}

func (s *Define) GetChildren() []Symbol {
	return s.children
}

func (s *Define) Consume(other Symbol) {
	if args, ok := other.(*ArgumentList); ok {
		firstArg := args.GetArguments()[0]
		if token, ok := firstArg.(*lexer.Token); ok {
			if token.Type == lexer.StringLiteral {
				stringText := util.GetTokenText(token, s.GetDocument().GetText())
				s.Name = string(stringText[1 : len(stringText)-1])
			}
		}
		if len(args.GetArguments()) >= 2 {
			secondArg := args.GetArguments()[1]
			s.Value = util.GetNodeText(secondArg, s.GetDocument().GetText())
		}
	}
}
