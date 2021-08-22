package wordtokeniser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var /* const */ docs map[string][]string = map[string][]string{
	"lowercase":           {"lowercase"},
	"Class":               {"Class"},
	"MyClass":             {"My", "Class"},
	"MyC":                 {"My", "C"},
	"HTML":                {"HTML"},
	"PDFLoader":           {"PDF", "Loader"},
	"AString":             {"A", "String"},
	"SimpleXMLParser":     {"Simple", "XML", "Parser"},
	"vimRPCPlugin":        {"vim", "RPC", "Plugin"},
	"GL11Version":         {"GL", "11", "Version"},
	"99Bottles":           {"99", "Bottles"},
	"May5":                {"May", "5"},
	"BFG9000":             {"BFG", "9000"},
	"BöseÜberraschung":    {"Böse", "Überraschung"},
	"BadUTF8\xe2\xe2\xa1": {"Bad", "UTF", "8"},
	"\xa7Filter":          {"Filter"},
}

var /* const */ docs2 map[string][]string = map[string][]string{
	"COMPLETION_COMPLETE": {"COMPLETION_COMPLETE", "COMPLETION", "COMPLETE"},
	"lowercase":           {"lowercase"},
	"Class":               {"Class"},
	"MyClass":             {"MyClass", "My", "Class"},
	"MyC":                 {"MyC", "My", "C"},
	"HTML":                {"HTML"},
	"PDFLoader":           {"PDFLoader", "PDF", "Loader"},
	"AString":             {"AString", "A", "String"},
	"SimpleXMLParser":     {"SimpleXMLParser", "Simple", "XML", "Parser"},
	"vimRPCPlugin":        {"vimRPCPlugin", "vim", "RPC", "Plugin"},
	"GL11Version":         {"GL11Version", "GL", "11", "Version"},
	"99Bottles":           {"99Bottles", "99", "Bottles"},
	"May5":                {"May5", "May", "5"},
	"BFG9000":             {"BFG9000", "BFG", "9000"},
	"BöseÜberraschung":    {"BöseÜberraschung", "Böse", "Überraschung"},
	"BadUTF8\xe2\xe2\xa1": {"BadUTF8\xe2\xe2\xa1", "Bad", "UTF", "8"},
	"\xa7Filter":          {"\xa7Filter", "Filter"},
	"MQF":                 {"MQF"},
	"Onedrive":            {"Onedrive"},
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

func BenchmarkTokenise(t *testing.B) {
	for i := 0; i < t.N; i++ {
		for doc := range docs2 {
			Tokenise(doc)
		}
	}
}

func TestTokenise(t *testing.T) {
	for doc, expected := range docs2 {
		actual := Tokenise(doc)
		assert.Equal(t, expected, actual)
	}
}
