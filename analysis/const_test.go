package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/indexer"
	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func TestConstant(t *testing.T) {
	constTest := "../cases/const.php"
	data, err := ioutil.ReadFile(constTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(constTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}

func TestConstantSerialiseAndDeserialise(t *testing.T) {
	constTest := "../cases/const.php"
	data, err := ioutil.ReadFile(constTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(constTest), string(data), rootNode)
	for _, child := range document.Children {
		if constant, ok := child.(*Const); ok {
			jsonData, _ := json.MarshalIndent(constant, "", "  ")
			original := string(jsonData)
			serialiser := indexer.NewSerialiser()
			constant.Serialise(serialiser)
			serialiser = indexer.SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedConstant := ReadConst(document, serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedConstant, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		} else if constant, ok := child.(*Define); ok {
			jsonData, _ := json.MarshalIndent(constant, "", "  ")
			original := string(jsonData)
			serialiser := indexer.NewSerialiser()
			constant.Serialise(serialiser)
			serialiser = indexer.SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedConstant := ReadDefine(document, serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedConstant, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
