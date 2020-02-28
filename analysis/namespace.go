package analysis

import (
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type Namespace struct {
	Name string
}

func newNamespace(document *Document, node *sitter.Node) *Namespace {
	namespace := &Namespace{}
	document.pushImportTable(node)
	document.setNamespace(namespace)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "namespace_name":
			namespace.Name = document.GetNodeText(child)
		case "compound_statement":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
	return namespace
}
