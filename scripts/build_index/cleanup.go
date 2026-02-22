package main

import "regexp"

var _parts = regexp.MustCompile(`\((Part|Subpart) [A-Z]{1,2}\)`)

// StripParts removes "(Part A)" and "(Subpart B) parenthentical content in text"
func StripParts(s string) string {
	return _parts.ReplaceAllString(s, "")
}
