package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestMultipleNamespaces(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/multipleNamespaces.php")
	assert.NoError(t, err)
	document := NewDocument("test1", data)
	document.Load()

	cupaloy.SnapshotT(t, document.Children)
	assert.Equal(t, "Namespace1", document.ImportTableAtPos(protocol.Position{
		Line:      3,
		Character: 0,
	}).GetNamespace())
	assert.Equal(t, "Namespace1", document.ImportTableAtPos(protocol.Position{
		Line:      8,
		Character: 0,
	}).GetNamespace())
	assert.Equal(t, "Namespace2", document.ImportTableAtPos(protocol.Position{
		Line:      9,
		Character: 22,
	}).GetNamespace())
}
