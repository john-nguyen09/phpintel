package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type AnonymousFunction struct {
	location protocol.Location

	Params []*Parameter
}

func newAnonymousFunction(document *Document, node *sitter.Node) Symbol {
	anonFunc := &AnonymousFunction{
		location: document.GetNodeLocation(node),
	}
	document.pushVariableTable(node)
	variableTable := document.getCurrentVariableTable()
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "formal_parameters":
			anonFunc.analyseParameterDeclarationList(document, child)
			for _, param := range anonFunc.Params {
				variableTable.add(param.ToVariable())
			}
			document.addSymbol(anonFunc)
		case "compound_statement":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
	document.popVariableTable()
	return nil
}

func (s *AnonymousFunction) GetLocation() protocol.Location {
	return s.location
}

func (s *AnonymousFunction) analyseParameterDeclarationList(document *Document, node *sitter.Node) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "simple_parameter":
			param := newParameter(document, child)
			s.Params = append(s.Params, param)
		}
		child = traverser.Advance()
	}
}
