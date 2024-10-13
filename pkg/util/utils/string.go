package utils

import (
	"unicode"

	"github.com/rivo/uniseg"
)

func IsStringAcceptable(str string) bool {
	for _, v := range str {
		if v == '\n' || v == '\r' || v == '\t' {
			continue
		}
		if !unicode.IsGraphic(v) {
			return false
		}
	}

	return true
}

func UnPrintable(s string) []rune {
	var unPrintable []rune

	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if !unicode.IsPrint(r) {
			unPrintable = append(unPrintable, r)
		}
	}

	return unPrintable
}

// CharacterCount counts one ASCII character is counted as one character, other characters are counted as two (including emoji).
func CharacterCount(str string) int {
	var count int

	grIterator := uniseg.NewGraphemes(str)
	for grIterator.Next() {
		r := grIterator.Runes()
		if len(r) > 1 || r[0] > unicode.MaxASCII {
			count += 2
		} else {
			count++
		}
	}

	return count
}
