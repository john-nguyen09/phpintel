package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

const source = "phpintel"

// GetParserDiagnostics returns the diagnostics for the syntax error
func GetParserDiagnostics(document *Document) []protocol.Diagnostic {
	rootNode := document.GetRootNode()
	diagnostics := []protocol.Diagnostic{}
	traverser := util.NewTraverser(rootNode)
	traverser.Traverse(func(node phrase.AstNode, _ []*phrase.Phrase) util.VisitorContext {
		if p, ok := node.(*phrase.Phrase); ok && p.Type == phrase.DocumentComment {
			return util.VisitorContext{ShouldAscend: false}
		}
		if err, ok := node.(*phrase.ParseError); ok {
			diagnostics = append(diagnostics, parserErrorToDiagnostic(document, err))
		}
		return util.VisitorContext{ShouldAscend: true}
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
		Source:   source,
	}
}

// UnusedDiagnostics returns the diagnostics for unused variables or imports
// TODO: provide unused imports
func UnusedDiagnostics(document *Document) []protocol.Diagnostic {
	diagnostics := []protocol.Diagnostic{}
	unusedVariables := document.UnusedVariables()
	for _, unusedVariable := range unusedVariables {
		nodes := document.NodeSpineAt(document.OffsetAtPosition(unusedVariable.Location.Range.End))
		nodes.Parent() // Ignore variable node
		parent := nodes.Parent()
		var r protocol.Range
		if parent.Type == phrase.SimpleAssignmentExpression {
			r = document.nodeRange(&parent)
		} else {
			r = unusedVariable.GetLocation().Range
		}
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range:    r,
			Message:  unusedVariable.Name + " is declared but its value is never read.",
			Source:   source,
			Severity: protocol.SeverityHint,
			Tags:     []protocol.DiagnosticTag{protocol.Unnecessary},
		})
	}
	return diagnostics
}
