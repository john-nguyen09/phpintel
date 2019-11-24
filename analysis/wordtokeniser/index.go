package wordtokeniser

import (
	"strings"
)

// Tokenise tokenises name into tokens for completion search
func Tokenise(name string) []string {
	lastSlashIndex := strings.LastIndex(name, "\\")
	if lastSlashIndex != -1 {
		name = string([]rune(name)[lastSlashIndex:])
	}

	// TODO: Combine underscore and casing tokenisers
	if strings.Contains(name, "_") {
		return underscore(name)
	} else {
		return casing(name)
	}
}
