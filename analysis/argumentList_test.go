package analysis

import (
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestNestedArgumentList(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/nestedArgs.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", string(data))
	document.Load()
	testOffsets := []int{
		308, 345,
	}
	for _, testOffset := range testOffsets {
		argumentList, hasParamsResolvable := document.ArgumentListAndFunctionCallAt(document.positionAt(testOffset))
		offsetStr := strconv.Itoa(testOffset)
		t.Run("TestNestedArgumentList"+offsetStr, func(t *testing.T) {
			cupaloy.SnapshotT(t, argumentList, hasParamsResolvable)
		})
	}
}
