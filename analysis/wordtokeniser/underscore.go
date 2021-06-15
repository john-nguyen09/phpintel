package wordtokeniser

func underscore(name string) []string {
	words := []string{}
	gotUnderscore := false
	start := 0
	end := len(name)
	for i, c := range name {
		if gotUnderscore {
			if c == '_' {
				continue
			}

			words = append(words, name[start:end])
			start = i
			gotUnderscore = false
		} else {
			if c == '_' {
				gotUnderscore = true
				end = i
			}
		}
	}
	if start < len(name) {
		words = append(words, name[start:])
	}
	return words
}
