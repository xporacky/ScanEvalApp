package common

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// removeDiacritics removes accents like é -> e, ň -> n, etc.
func RemoveDiacritics(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isDiacritic), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// checks for diacritic
func isDiacritic(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

// sanitizeFilename converts string to safe ASCII-only filename
func SanitizeFilename(name string) string {
	// Remove diacritics
	name = RemoveDiacritics(name)

	// Replace spaces and dashes with underscores
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Remove all non-alphanumeric, non-underscore characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	name = reg.ReplaceAllString(name, "")

	return name
}
