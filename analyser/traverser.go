package analyser

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
)

func Traverse(node phrase.AstNode, analyser *Analyser) {
	if p, ok := node.(*phrase.Phrase); ok {
		analyser.Preorder(p)

		for _, child := range p.Children {
			Traverse(child, analyser)
		}

		analyser.Postorder(p)
	}
}
