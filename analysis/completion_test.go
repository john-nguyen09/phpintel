package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func indexTestCase(store *Store, uri string, path string, isOpen bool) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	document := NewDocument(uri, string(data))
	document.Load()
	document.isOpen = isOpen
	store.SyncDocument(document)
}

func TestFunctionCompletionWithNamespace(t *testing.T) {
	store, err := setupStore("test", "TestFunctionCompletionWithNamespace")
	if err != nil {
		panic(err)
	}
	indexTestCase(store, "test1", "../cases/function.php", false)
	indexTestCase(store, "test2", "../cases/completion/functionCompletionWithNamespace.php", true)
	document := store.GetOrCreateDocument("test2")
	importTable := document.GetImportTable()
	word := "testF"
	functions := store.SearchFunctions(word)
	items := []protocol.CompletionItem{}
	for _, function := range functions {
		label, textEdit := importTable.ResolveToQualified(document, function, function.Name, word)
		textEdits := []protocol.TextEdit{}
		if textEdit != nil {
			textEdits = append(textEdits, *textEdit)
		}
		items = append(items, protocol.CompletionItem{
			Kind:                protocol.FunctionCompletion,
			Label:               label,
			AdditionalTextEdits: textEdits,
		})
	}
	cupaloy.SnapshotT(t, items)
}