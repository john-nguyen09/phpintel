package php

//#include "parser.h"
//TSLanguage *tree_sitter_php();
import "C"
import (
	"unsafe"

	sitter "github.com/smacker/go-tree-sitter"
)

var injectionQuery []byte = []byte(`((comment) @injection.content
 (set! injection.regex "^/\*\*")
 (set! injection.language "phpdoc"))`)

func GetLanguage() *sitter.Language {
	ptr := unsafe.Pointer(C.tree_sitter_php())
	return sitter.NewLanguage(ptr)
}

func GetInjectionQuery() []byte {
	return injectionQuery
}
