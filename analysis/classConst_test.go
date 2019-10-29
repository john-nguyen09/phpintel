package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestClassConst(t *testing.T) {
	classConstTest := "../cases/classConst.php"
	data, err := ioutil.ReadFile(classConstTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(classConstTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}

func TestClassConstSerialiseAndDeserialise(t *testing.T) {
	classConstTest := "../cases/classConst.php"
	data, err := ioutil.ReadFile(classConstTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(classConstTest), string(data), rootNode)
	for _, child := range document.Children {
		if classConst, ok := child.(*ClassConst); ok {
			jsonData, _ := json.MarshalIndent(classConst, "", "  ")
			original := string(jsonData)
			serialiser := NewSerialiser()
			classConst.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedClassConst := ReadClassConst(serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedClassConst, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
