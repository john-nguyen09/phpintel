package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

type Namespace struct {
	Name TypeString
}

func newNamespace(document *Document, node *phrase.Phrase) *Namespace {
	namespace := &Namespace{}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.NamespaceName:
				namespace.Name = newTypeString("\\" + document.GetPhraseText(node))
			}
		}
		child = traverser.Advance()
	}
	return namespace
}
