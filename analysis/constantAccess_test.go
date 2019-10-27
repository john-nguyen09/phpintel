package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func TestConstantAccess(t *testing.T) {
	constantAccessTest := "../cases/constantAccess.php"
	data, err := ioutil.ReadFile(constantAccessTest)
	if err != nil {
		panic(err)
	}
	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(constantAccessTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}

func TestConstantAccessSerialiseAndDeserialise(t *testing.T) {
	constantAccessTest := "../cases/constantAccess.php"
	data, err := ioutil.ReadFile(constantAccessTest)
	if err != nil {
		panic(err)
	}
	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(constantAccessTest), string(data), rootNode)
	for _, child := range document.Children {
		if constantAccess, ok := child.(*ConstantAccess); ok {
			jsonData, _ := json.MarshalIndent(constantAccess, "", "  ")
			original := string(jsonData)
			serialiser := NewSerialiser()
			constantAccess.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedConstantAccess := ReadConstantAccess(serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedConstantAccess, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
