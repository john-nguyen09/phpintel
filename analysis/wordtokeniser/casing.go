package wordtokeniser

import (
	"unicode"
	"unicode/utf8"
)

func casing(name string) []string {
	tokens := []string{}
	lastClass := 0
	start := 0
	nameRunes := []rune(name)

	tokens = append(tokens, name)
	for i, r := range nameRunes {
		if r == utf8.RuneError {
			if start > 0 && start < i {
				tokens = append(tokens, string(nameRunes[start:i]))
			}
			start = i + 1
			continue
		}
		class := 1
		switch true {
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		}
		if i != 0 && class != lastClass {
			// It is acceptable when going from UPPERCASE to lowercase for 1 character
			// e.g. Class -> ["Class"] instead of ["C", "lass"]
			// But ABClass -> ["AB", "Class"], instead of ["ABC", "lass"] or ["ABClass"]
			if lastClass == 2 && class == 1 {
				if start != i-1 {
					tokens = append(tokens, string(nameRunes[start:i-1]))
				}
				start = i - 1
			} else {
				tokens = append(tokens, string(nameRunes[start:i]))
				start = i
			}
		}
		lastClass = class
	}
	if start > 0 && start < len(nameRunes) {
		tokens = append(tokens, string(nameRunes[start:]))
	}

	return tokens
}
