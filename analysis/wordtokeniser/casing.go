package wordtokeniser

import (
	"log"
	"unicode"
	"unicode/utf8"
)

func isNotToken(r rune) bool {
	return r == '_'
}

func casing(name string) []string {
	words := []string{}
	lastClass := 0
	start := 0
	prev := 0
	lastTokenIndex := 0
	markNextAsStart := false
	appendWord := func(word string) {
		if start == 0 {
			return
		}
		words = append(words, word)
	}
	for i, r := range name {
		if isNotToken(r) {
			if start < lastTokenIndex {
				words = append(words, name[start:lastTokenIndex+1])
			}
			lastClass = 0
			start = i + 1
			prev = i + 1
			continue
		}
		lastTokenIndex = i
		if markNextAsStart {
			start = i
			markNextAsStart = false
		}
		if r == utf8.RuneError {
			if start > 0 && start < i {
				appendWord(name[start:i])
			}
			markNextAsStart = true
			if i >= len(name)-1 {
				start = len(name)
			}
			continue
		}
		class := 1
		switch {
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		}
		if lastClass != 0 && class != lastClass {
			// It is acceptable when going from UPPERCASE to lowercase for 1 character
			// e.g. Class -> ["Class"] instead of ["C", "lass"]
			// But ABClass -> ["AB", "Class"], instead of ["ABC", "lass"] or ["ABClass"]
			if lastClass == 2 && class == 1 {
				if start != prev {
					appendWord(name[start:prev])
				}
				start = prev
			} else {
				if start > i {
					log.Printf("start: %d, i: %d, name: %s, bytes: %v", start, i, name, []byte(name))
				}
				appendWord(name[start:i])
				start = i
			}
		}
		lastClass = class
		prev = i
	}
	if start < len(name) {
		appendWord(string(name[start:]))
	}
	return words
}
