package validation

import (
	"unicode"
)

func HasValidCharacters(sn string) bool {
	var hasLetter bool
	for _, r := range sn {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return false
	}

	if len(sn) > 8 {
		return false
	}

	for _, r := range sn {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-' {
			return false
		}
	}
	return true
}
