package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// normalize converts a string to lowercase, removes accents, and strips symbols.
func Normalize(input string) string {
	// Step 1: lowercase
	input = strings.ToLower(input)

	// Step 2: remove diacritics (é -> e, ü -> u, etc.)
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn = nonspacing marks
	}), norm.NFC)

	result, _, _ := transform.String(t, input)

	// Step 3: remove symbols (keep letters, numbers, spaces)
	reg := regexp.MustCompile(`[^a-z0-9\s]+`)
	result = reg.ReplaceAllString(result, "")

	// Step 4: trim multiple spaces
	result = strings.Join(strings.Fields(result), " ")

	return result
}
