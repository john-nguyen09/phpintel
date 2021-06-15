package wordtokeniser

import (
	"log"
	"unicode"
	"unicode/utf8"
)

func casing(name string) []string {
	words := []string{}
	lastClass := 0
	start := 0
	prev := 0
	markNextAsStart := false
	isBegin := true
	for i, r := range name {
		if markNextAsStart {
			start = i
			markNextAsStart = false
		}
		if r == utf8.RuneError {
			if start > 0 && start < i {
				words = append(words, name[start:i])
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
		if !isBegin && class != lastClass {
			// It is acceptable when going from UPPERCASE to lowercase for 1 character
			// e.g. Class -> ["Class"] instead of ["C", "lass"]
			// But ABClass -> ["AB", "Class"], instead of ["ABC", "lass"] or ["ABClass"]
			if lastClass == 2 && class == 1 {
				if start != prev {
					words = append(words, name[start:prev])
				}
				start = prev
			} else {
				if start > i {
					log.Printf("start: %d, i: %d, name: %s, bytes: %v", start, i, name, []byte(name))
				}
				words = append(words, name[start:i])
				start = i
			}
		}
		lastClass = class
		prev = i
		isBegin = false
	}
	if start < len(name) {
		words = append(words, string(name[start:]))
	}
	return words
}
