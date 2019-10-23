package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

func readMemberName(document *Document, node *phrase.Phrase) string {
	return util.GetNodeText(node, document.GetText())
}
