package analysis

import (
	"io/ioutil"
	"strconv"
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

type indexableNamespaceResult struct {
	is  []*indexableNamespace
	key string
}

type indexableNamspaceTestCase struct {
	namespaceName string
	expected      indexableNamespaceResult
}

func TestIndexableNamespace(t *testing.T) {
	testCases := []indexableNamspaceTestCase{
		{"", indexableNamespaceResult{[]*indexableNamespace{}, "\\"}},
		{"\\", indexableNamespaceResult{[]*indexableNamespace{}, "\\"}},
		{"TestNamespace1", indexableNamespaceResult{[]*indexableNamespace{
			{scope: "", name: "TestNamespace1"},
		}, "\\TestNamespace1"}},
		{"\\TestNamespace1", indexableNamespaceResult{[]*indexableNamespace{
			{scope: "", name: "TestNamespace1"},
		}, "\\TestNamespace1"}},
		{"Namespace1\\Namespace2", indexableNamespaceResult{[]*indexableNamespace{
			{scope: "", name: "Namespace1"},
			{scope: "Namespace1", name: "Namespace2"},
		}, "\\Namespace1\\Namespace2"}},
		{"\\Namespace1\\Namespace2", indexableNamespaceResult{[]*indexableNamespace{
			{scope: "", name: "Namespace1"},
			{scope: "Namespace1", name: "Namespace2"},
		}, "\\Namespace1\\Namespace2"}},
		{"Namespace1\\Namespace2\\Namespace3", indexableNamespaceResult{[]*indexableNamespace{
			{scope: "", name: "Namespace1"},
			{scope: "Namespace1", name: "Namespace2"},
			{scope: "Namespace1\\Namespace2", name: "Namespace3"},
		}, "\\Namespace1\\Namespace2\\Namespace3"}},
		{"\\Namespace1\\Namespace2\\Namespace3", indexableNamespaceResult{[]*indexableNamespace{
			{scope: "", name: "Namespace1"},
			{scope: "Namespace1", name: "Namespace2"},
			{scope: "Namespace1\\Namespace2", name: "Namespace3"},
		}, "\\Namespace1\\Namespace2\\Namespace3"}},
	}
	for i, testCase := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual, key := indexablesFromNamespaceName(testCase.namespaceName)
			assert.Equal(t, testCase.expected, indexableNamespaceResult{actual, key})
		})
	}
}
