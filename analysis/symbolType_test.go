package analysis

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestTypeString(t *testing.T) {
	testCases := []string{
		"string",
		"string[]",
		"string[][]",
		"int",
		"int[]",
	}
	results := []TypeString{}
	for _, testCase := range testCases {
		results = append(results, NewTypeString(testCase))
	}
	cupaloy.SnapshotT(t, results)
}
