package entity

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

func GetNames(p *phrase.Phrase, doc *PhpDoc) []string {
	names := make([]string, 0)

	for _, child := range p.Children {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedName {
			names = append(names, string(util.GetPhraseText(p, doc.Text)))
		}
	}

	return names
}
