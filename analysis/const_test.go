package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
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
			bytes := constant.Serialise()
			jsonData, _ := json.MarshalIndent(constant, "", "  ")
			original := string(jsonData)
			deserialisedConstant := DeserialiseConst(document, bytes)
			jsonData, _ = json.MarshalIndent(deserialisedConstant, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		} else if constant, ok := child.(*Define); ok {
			bytes := constant.Serialise()
			jsonData, _ := json.MarshalIndent(constant, "", "  ")
			original := string(jsonData)
			deserialisedConstant := DeserialiseDefine(document, bytes)
			jsonData, _ = json.MarshalIndent(deserialisedConstant, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
