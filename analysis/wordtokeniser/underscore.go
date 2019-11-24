package wordtokeniser

func underscore(name string) []string {
	tokens := []string{}
	gotUnderscore := false
	nameRunes := []rune(name)

	tokens = append(tokens, name)
	for i, r := range nameRunes {
		if gotUnderscore {
			if r == '_' {
				continue
			}

			tokens = append(tokens, string(nameRunes[i:]))
			gotUnderscore = false
		} else {
			if r == '_' {
				gotUnderscore = true
			}
		}
	}

	return tokens
}
