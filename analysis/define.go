package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Define contains information of define constants
type Define struct {
	document *Document
	location lsp.Location
	children []Symbol

	Name  string
	Value string
}

func newDefine(document *Document, parent symbolBlock, node *phrase.Phrase) Symbol {
	define := &Define{
		document: document,
		location: document.GetNodeLocation(node),
	}

	scanForChildren(define, node)

	return define
}

func (s *Define) getLocation() lsp.Location {
	return s.location
}

func (s *Define) getDocument() *Document {
	return s.document
}

func (s *Define) consume(other Symbol) {
	if args, ok := other.(*ArgumentList); ok {
		firstArg := args.GetArguments()[0]
		if token, ok := firstArg.(*lexer.Token); ok {
			if token.Type == lexer.StringLiteral {
				stringText := util.GetTokenText(token, s.getDocument().GetText())
				s.Name = string(stringText[1 : len(stringText)-1])
			}
		}
		if len(args.GetArguments()) >= 2 {
			secondArg := args.GetArguments()[1]
			s.Value = util.GetNodeText(secondArg, s.getDocument().GetText())
		}
	}
}
