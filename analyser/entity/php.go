package entity

import (
	"io/ioutil"

	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

type PhpDoc struct {
	Uri     string
	Text    []rune
	Classes []*Class
}

func NewPhpDoc(uri string) *PhpDoc {
	filePath := util.UriToPath(uri)
	byteBuffer, err := ioutil.ReadFile(filePath)

	if err != nil {
		util.HandleError(err)
	}

	return &PhpDoc{
		Uri:  uri,
		Text: []rune(string(byteBuffer))}
}

func (doc *PhpDoc) ParseAST() *phrase.Phrase {
	return parser.Parse(string(doc.Text))
}

func (doc *PhpDoc) AddClass(class *Class) {
	doc.Classes = append(doc.Classes, class)
}
