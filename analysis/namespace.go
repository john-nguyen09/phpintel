package analysis

import (
	"strings"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

type Namespace struct {
	Name string
}

func newNamespace(a analyser, document *Document, node *phrase.Phrase) *Namespace {
	namespace := &Namespace{}
	document.pushImportTable(node)
	document.setNamespace(namespace)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.NamespaceName:
				namespace.Name = document.getPhraseText(p)
			case phrase.StatementList:
				scanForChildren(a, document, p)
			}
		}
		child = traverser.Advance()
	}
	return namespace
}

type indexableNamespace struct {
	// The scope of the namespace
	scope string
	// The current name of a part
	name string
}

var _ NameIndexable = (*indexableNamespace)(nil)

func indexablesFromNamespaceName(namespaceName string) ([]*indexableNamespace, string) {
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
			})
			if scope != "" {
				scope += "\\"
			}
			scope += part
		}
	}
	return is, "\\" + name
}

func (i *indexableNamespace) GetIndexableName() string {
	return i.name
}

func (i *indexableNamespace) GetIndexCollection() string {
	scope := i.scope
	if len(scope) == 0 || scope == "\\" {
		return namespaceCompletionIndex
	}
	scope = "\\" + scope
	return namespaceCompletionIndex + KeySep + scope
}
