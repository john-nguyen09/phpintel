package analysis

import (
	"strings"

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
			namespace.Name = readNamespaceExcludeError(document, child)
		case "compound_statement":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
	return namespace
}

func readNamespaceExcludeError(document *Document, node *sitter.Node) string {
	traverser := util.NewTraverser(node)
	parts := []string{}
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "name":
			parts = append(parts, document.GetNodeText(child))
		}
	}
	return strings.Join(parts, "\\")
}

type indexableNamespace struct {
	// The scope of the namespace
	scope string
	// The current name of a part
	name string
	// key is the namespaceName
	key string
}

var _ NameIndexable = (*indexableNamespace)(nil)

func indexablesFromNamespaceName(namespaceName string) []*indexableNamespace {
	is := []*indexableNamespace{}
	name := namespaceName
	if len(name) > 0 && name[0] == '\\' {
		name = name[1:]
	}
	// Empty namespaces don't need index
	if len(name) > 0 {
		parts := strings.Split(name, "\\")
		scope := ""
		for _, part := range parts {
			is = append(is, &indexableNamespace{
				scope: scope,
				name:  part,
				key:   "\\" + name,
			})
			if scope != "" {
				scope += "\\"
			}
			scope += part
		}
	}
	return is
}

func (i *indexableNamespace) GetIndexableName() string {
	return i.name
}

func (i *indexableNamespace) GetIndexCollection() string {
	scope := i.scope
	if len(scope) > 0 && scope != "\\" {
		scope = "\\" + scope
	}
	return namespaceCompletionIndex + KeySep + scope
}
