package analysis

import (
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestNestedArgumentList(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/nestedArgs.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	testOffsets := []int{
		308,
		345,
	}
	for _, testOffset := range testOffsets {
		argumentList, hasParamsResolvable := document.ArgumentListAndFunctionCallAt(document.positionAt(testOffset))
		t.Run(strconv.Itoa(testOffset), func(t *testing.T) {
			cupaloy.SnapshotT(t, argumentList, hasParamsResolvable)
		})
	}
}

func TestNotDuplicatedExpression(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/argumentsExpression.php")
	assert.NoError(t, err)
	document := NewDocument("test1", data)
	document.Load()

	cupaloy.SnapshotT(t, document.hasTypesSymbols())
}
