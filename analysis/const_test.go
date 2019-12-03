package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestConstant(t *testing.T) {
	constTest := "../cases/const.php"
	data, err := ioutil.ReadFile(constTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestConstantSerialiseAndDeserialise(t *testing.T) {
	constTest := "../cases/const.php"
	data, err := ioutil.ReadFile(constTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	for _, child := range document.Children {
		if constant, ok := child.(*Const); ok {
			jsonData, _ := json.MarshalIndent(constant, "", "  ")
			original := string(jsonData)
			serialiser := NewSerialiser()
			constant.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedConstant := ReadConst(serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedConstant, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		} else if constant, ok := child.(*Define); ok {
			jsonData, _ := json.MarshalIndent(constant, "", "  ")
			original := string(jsonData)
			serialiser := NewSerialiser()
			constant.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedConstant := ReadDefine(serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedConstant, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
