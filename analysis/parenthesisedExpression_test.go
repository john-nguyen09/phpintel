package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestParenthesisedExpression(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/parenthesised.php")
	assert.NoError(t, err)
	doc := NewDocument("test1", data)
	doc.Load()

	cupaloy.SnapshotT(t, doc.hasTypesSymbols)
}
