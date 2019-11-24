package wordtokeniser

import "unicode"

func casing(name string) []string {
	tokens := []string{}
	isPrevUpper := false
	start := -1
	nameRunes := []rune(name)

	tokens = append(tokens, name)
	for i, r := range nameRunes {
		isCurrUpper := unicode.IsUpper(r)

		if isCurrUpper != isPrevUpper {
			if start == -1 {
				if isCurrUpper {
					if i != 0 {
						start = i
					}
				} else {
					if i != 1 {
						tokens = append(tokens, string(nameRunes[i-1:]))
					}
				}
			} else {
				tokens = append(tokens, string(nameRunes[start:]))
				if start != (i-1) && !isCurrUpper {
					tokens = append(tokens, string(nameRunes[i-1:]))
				}
				start = -1
			}
		}

		isPrevUpper = isCurrUpper
	}

	return tokens
}
