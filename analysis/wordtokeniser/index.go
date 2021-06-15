package wordtokeniser

import (
	"sort"
)

// Tokenise tokenises name into tokens for completion search
func Tokenise(name string) []string {
	words := underscore(name)
	results := []string{
		name,
	}
	for _, word := range words {
		results = append(results, casing(word)...)
	}
	return removeDup(results)
}

func removeDup(in []string) []string {
	sort.Strings(in)
	j := 0
	for i := 1; i < len(in); i++ {
		if in[j] == in[i] {
			continue
		}
		j++
		in[j] = in[i]
	}
	return in[:j+1]
}
