package analyser

import (
	"path"
	"runtime"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/phpintel/analyser/entity"
)

func indexFiles(filePaths []string) []*entity.PhpDoc {
	phpDocs := make([]*entity.PhpDoc, 0)

	for _, filePath := range filePaths {
		phpDoc := entity.NewPhpDoc(util.PathToUri(filePath))
		analyser := NewAnalyser(phpDoc)
		Traverse(phpDoc.ParseAST(), analyser)

		phpDocs = append(phpDocs, phpDoc)
	}

	return phpDocs
}

func TestAnalyseClass(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	currentDir := path.Dir(filename)

	phpDocs := indexFiles([]string{
		path.Clean(currentDir + "/../case/class.php")})

	for _, phpDoc := range phpDocs {
		spew.Dump(phpDoc.Classes)
	}
}
