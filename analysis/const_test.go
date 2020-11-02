package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
)

func TestConstant(t *testing.T) {
	constTest := "../cases/const.php"
	data, err := ioutil.ReadFile(constTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	results := []Symbol{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol, _ []Symbol) {
		switch s.(type) {
		case *Const, *Define:
			results = append(results, s)
		}
	})
	cupaloy.SnapshotT(t, results)
}

func TestConstantSerialiseAndDeserialise(t *testing.T) {
	constTest := "../cases/const.php"
	data, err := ioutil.ReadFile(constTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	for _, child := range document.Children {
		if constant, ok := child.(*Const); ok {
			jsonData, _ := json.MarshalIndent(constant, "", "  ")
			original := string(jsonData)
			e := storage.NewEncoder()
			constant.Serialise(e)
			d := storage.NewDecoder(e.Bytes())
			deserialisedConstant := ReadConst(d)
			jsonData, _ = json.MarshalIndent(deserialisedConstant, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		} else if constant, ok := child.(*Define); ok {
			jsonData, _ := json.MarshalIndent(constant, "", "  ")
			original := string(jsonData)
			e := storage.NewEncoder()
			constant.Serialise(e)
			d := storage.NewDecoder(e.Bytes())
			deserialisedConstant := ReadDefine(d)
			jsonData, _ = json.MarshalIndent(deserialisedConstant, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
