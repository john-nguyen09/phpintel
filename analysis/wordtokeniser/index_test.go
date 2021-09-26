package wordtokeniser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var /* const */ docs map[string][]string = map[string][]string{
	"COMPLETION_COMPLETE": {"COMPLETION_COMPLETE", "COMPLETE"},
	"_Name1":              {"_Name1", "Name1", "Name", "1"},
	"lowercase":           {"lowercase"},
	"Class":               {"Class"},
	"MyClass":             {"MyClass", "Class"},
	"MyC":                 {"MyC", "C"},
	"HTML":                {"HTML"},
	"PDFLoader":           {"PDFLoader", "Loader"},
	"AString":             {"AString", "String"},
	"SimpleXMLParser":     {"SimpleXMLParser", "XML", "Parser"},
	"vimRPCPlugin":        {"vimRPCPlugin", "RPC", "Plugin"},
	"GL11Version":         {"GL11Version", "11", "Version"},
	"99Bottles":           {"99Bottles", "Bottles"},
	"May5":                {"May5", "5"},
	"BFG9000":             {"BFG9000", "9000"},
	"BöseÜberraschung":    {"BöseÜberraschung", "Überraschung"},
	"BadUTF8\xe2\xe2\xa1": {"BadUTF8\xe2\xe2\xa1", "UTF", "8"},
	"\xa7Filter":          {"\xa7Filter", "Filter"},
	"MQF":                 {"MQF"},
	"Onedrive":            {"Onedrive"},
}

func BenchmarkTokenise(t *testing.B) {
	for i := 0; i < t.N; i++ {
		for doc := range docs {
			Tokenise(doc)
		}
	}
}

func TestTokenise(t *testing.T) {
	for doc, expected := range docs {
		actual := Tokenise(doc)
		assert.Equal(t, expected, actual)
	}
}
