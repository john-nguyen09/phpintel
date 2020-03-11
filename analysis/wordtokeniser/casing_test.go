package wordtokeniser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var /* const */ docs map[string][]string = map[string][]string{
	"lowercase":           []string{"lowercase"},
	"Class":               []string{"Class"},
	"MyClass":             []string{"MyClass", "My", "Class"},
	"MyC":                 []string{"MyC", "My", "C"},
	"HTML":                []string{"HTML"},
	"PDFLoader":           []string{"PDFLoader", "PDF", "Loader"},
	"AString":             []string{"AString", "A", "String"},
	"SimpleXMLParser":     []string{"SimpleXMLParser", "Simple", "XML", "Parser"},
	"vimRPCPlugin":        []string{"vimRPCPlugin", "vim", "RPC", "Plugin"},
	"GL11Version":         []string{"GL11Version", "GL", "11", "Version"},
	"99Bottles":           []string{"99Bottles", "99", "Bottles"},
	"May5":                []string{"May5", "May", "5"},
	"BFG9000":             []string{"BFG9000", "BFG", "9000"},
	"BöseÜberraschung":    []string{"BöseÜberraschung", "Böse", "Überraschung"},
	"BadUTF8\xe2\xe2\xa1": []string{"BadUTF8\xe2\xe2\xa1", "Bad", "UTF", "8"},
}

func BenchmarkCasing(t *testing.B) {
	for i := 0; i < t.N; i++ {
		for doc := range docs {
			casing(doc)
		}
	}
}

func TestCasing(t *testing.T) {
	for doc, expected := range docs {
		actual := casing(doc)
		assert.Equal(t, expected, actual)
	}
}
