package wordtokeniser

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
