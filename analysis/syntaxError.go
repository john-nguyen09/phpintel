package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

func GetParserDiagnostic(document *Document) []protocol.Diagnostic {
	rootNode := document.GetRootNode()
	diagnostics := []protocol.Diagnostic{}
	traverser := util.NewTraverser(rootNode)
	traverser.Traverse(func(node *sitter.Node, _ []*sitter.Node) util.VisitorContext {
		t := node.Type()
		switch {
		case t == "ERROR" || node.IsMissing():
			diagnostics = append(diagnostics, parserErrorToDiagnostic(document, node))
		}
		return util.VisitorContext{true, nil}
	})

	return diagnostics
}

func parserErrorToDiagnostic(document *Document, err *sitter.Node) protocol.Diagnostic {
	message := err.Type() + "."
	if err.IsMissing() {
		message = "Missing: " + err.Type() + "."
	}

	return protocol.Diagnostic{
		Range:    document.nodeRange(err),
		Message:  message,
		Severity: protocol.SeverityError,
		Source:   "phpintel",
	}
}
