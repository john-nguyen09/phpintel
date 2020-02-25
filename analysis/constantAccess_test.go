package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
)

func TestConstantAccess(t *testing.T) {
	constantAccessTest := "../cases/constantAccess.php"
	data, err := ioutil.ReadFile(constantAccessTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestConstantAccessSerialiseAndDeserialise(t *testing.T) {
	constantAccessTest := "../cases/constantAccess.php"
	data, err := ioutil.ReadFile(constantAccessTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	for _, child := range document.Children {
		if constantAccess, ok := child.(*ConstantAccess); ok {
			jsonData, _ := json.MarshalIndent(constantAccess, "", "  ")
			original := string(jsonData)
			e := storage.NewEncoder()
			constantAccess.Serialise(e)
			d := storage.NewDecoder(e.Bytes())
			deserialisedConstantAccess := ReadConstantAccess(d)
			jsonData, _ = json.MarshalIndent(deserialisedConstantAccess, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
