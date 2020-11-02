package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
)

func TestClassConst(t *testing.T) {
	classConstTest := "../cases/classConst.php"
	data, err := ioutil.ReadFile(classConstTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	results := []Symbol{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol, _ []Symbol) {
		if _, ok := s.(*ClassConst); ok {
			results = append(results, s)
		}
	})
	cupaloy.SnapshotT(t, results)
}

func TestClassConstSerialiseAndDeserialise(t *testing.T) {
	classConstTest := "../cases/classConst.php"
	data, err := ioutil.ReadFile(classConstTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	for _, child := range document.Children {
		if classConst, ok := child.(*ClassConst); ok {
			jsonData, _ := json.MarshalIndent(classConst, "", "  ")
			original := string(jsonData)
			e := storage.NewEncoder()
			classConst.Serialise(e)
			d := storage.NewDecoder(e.Bytes())
			deserialisedClassConst := ReadClassConst(d)
			jsonData, _ = json.MarshalIndent(deserialisedClassConst, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
