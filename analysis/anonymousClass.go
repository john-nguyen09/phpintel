package analysis

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

func newAnonymousClass(a analyser, document *Document, node *phrase.Phrase) Symbol {
	class := &Class{
		Location: document.GetNodeLocation(node),
	}
	hash := md5.Sum([]byte(class.Location.URI))
	class.Name = NewTypeString("anonClass#" + hex.EncodeToString(hash[:6]) + "#" + class.Location.Range.String())
	document.addClass(class)
	phpDoc := document.getValidPhpDoc(class.Location)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	document.addSymbol(class)
	document.pushBlock(class)

	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.AnonymousClassDeclarationHeader:
				class.analyseHeader(a, document, p, phpDoc)
			case phrase.ClassDeclarationBody:
				scanForChildren(a, document, p)
			}
		}
		child = traverser.Advance()
	}

	document.popBlock()

	return nil
}
