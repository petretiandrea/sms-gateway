package config

import "strings"

func StripUnderscore(s, newDelimiter string) string {
	var new = strings.Replace(s, "__", "@", -1)
	new = strings.Replace(new, "_", newDelimiter, -1)
	new = strings.Replace(new, "@", "_", -1)
	return new
}