package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func GetParserDiagnostic(document *Document) []protocol.Diagnostic {
	rootNode := document.GetRootNode()
	diagnostics := []protocol.Diagnostic{}
	traverser := util.NewTraverser(rootNode)
	traverser.Traverse(func(node phrase.AstNode) bool {
		if err, ok := node.(*phrase.ParseError); ok {
			diagnostics = append(diagnostics, parserErrorToDiagnostic(document, err))
		}
		return true
	})

	return diagnostics
}

func parserErrorToDiagnostic(document *Document, err *phrase.ParseError) protocol.Diagnostic {
	message := "Unexpected " + err.Type.String() + "."
	if err.Expected != lexer.Undefined {
		message += " Expected " + err.Expected.String() + "."
	}

	return protocol.Diagnostic{
		Range:    document.errorRange(err),
		Message:  message,
		Severity: protocol.SeverityError,
		Source:   "phpintel",
	}
}
