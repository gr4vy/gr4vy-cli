package main

import "unicode"

// kebab converts a Pascal/camelCase identifier to kebab-case, handling acronym
// runs ("merchantAccountID" -> "merchant-account-id", "URL" -> "url").
func kebab(s string) string {
	rs := []rune(s)
	var b []rune
	for i, r := range rs {
		if i > 0 && unicode.IsUpper(r) {
			prev := rs[i-1]
			var next rune
			if i+1 < len(rs) {
				next = rs[i+1]
			}
			if unicode.IsLower(prev) || unicode.IsDigit(prev) ||
				(unicode.IsUpper(prev) && next != 0 && unicode.IsLower(next)) {
				b = append(b, '-')
			}
		}
		b = append(b, unicode.ToLower(r))
	}
	return string(b)
}

// pascalToKebab is kebab for resource field names.
func pascalToKebab(s string) string { return kebab(s) }

// capitalize upper-cases the first rune of s (a replacement for the deprecated
// strings.Title, sufficient for single-word command verbs).
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// snakeToKebab converts snake_case (from spec/tag names) to kebab-case.
func snakeToKebab(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '_' {
			out = append(out, '-')
		} else {
			out = append(out, unicode.ToLower(r))
		}
	}
	return string(out)
}
