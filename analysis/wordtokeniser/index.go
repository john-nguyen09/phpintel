package wordtokeniser

import "strings"

// Tokenise tokenises name into tokens for completion search
func Tokenise(name string) []string {
	results := []string{
		name,
	}
	nameTrimmed := strings.Trim(name, "_")
	if nameTrimmed != name {
		results = append(results, nameTrimmed)
	}
	results = append(results, casing(name)...)
	return removeDup(results)
}

type void struct{}

var empty void = void{}

func removeDup(in []string) []string {
	out := []string{}
	set := map[string]void{}
	for _, str := range in {
		if _, ok := set[str]; ok {
			continue
		}
		out = append(out, str)
		set[str] = empty
	}
	return out
}
