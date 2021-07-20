package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestDeprecatedReferences(t *testing.T) {
	withTestStore("test", t.Name(), func(store *Store) {
		indexDocument(store, "../cases/deprecated/definitions.php", "definitions")
		data, err := ioutil.ReadFile("../cases/deprecated/references.php")
		if err != nil {
			panic(err)
		}
		doc := NewDocument("references", data)
		doc.Load()
		store.SyncDocument(doc)

		diagnostics := DeprecatedDiagnostics(NewResolveContext(NewQuery(store), doc))
		results := []protocol.Range{}
		for _, diagnostic := range diagnostics {
			results = append(results, diagnostic.Range)
		}
		assert.Equal(t, []protocol.Range{
			{Start: protocol.Position{Line: 2, Character: 0}, End: protocol.Position{Line: 2, Character: 18}},
			{Start: protocol.Position{Line: 2, Character: 19}, End: protocol.Position{Line: 2, Character: 35}},
			{Start: protocol.Position{Line: 4, Character: 23}, End: protocol.Position{Line: 4, Character: 38}},
			{Start: protocol.Position{Line: 4, Character: 39}, End: protocol.Position{Line: 4, Character: 87}},
			{Start: protocol.Position{Line: 5, Character: 18}, End: protocol.Position{Line: 5, Character: 33}},
			{Start: protocol.Position{Line: 5, Character: 36}, End: protocol.Position{Line: 5, Character: 51}},
			{Start: protocol.Position{Line: 5, Character: 53}, End: protocol.Position{Line: 5, Character: 75}},
			{Start: protocol.Position{Line: 6, Character: 18}, End: protocol.Position{Line: 6, Character: 34}},
			{Start: protocol.Position{Line: 7, Character: 0}, End: protocol.Position{Line: 7, Character: 15}},
			{Start: protocol.Position{Line: 7, Character: 17}, End: protocol.Position{Line: 7, Character: 39}},
			{Start: protocol.Position{Line: 8, Character: 0}, End: protocol.Position{Line: 8, Character: 15}},
			{Start: protocol.Position{Line: 8, Character: 17}, End: protocol.Position{Line: 8, Character: 38}},
			{Start: protocol.Position{Line: 10, Character: 40}, End: protocol.Position{Line: 10, Character: 55}},
			{Start: protocol.Position{Line: 10, Character: 67}, End: protocol.Position{Line: 10, Character: 86}},
			{Start: protocol.Position{Line: 12, Character: 29}, End: protocol.Position{Line: 12, Character: 44}},
		}, results)
	})
}
