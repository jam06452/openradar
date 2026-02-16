// Groq api keys
package detectors

import (
	"regexp"
)

var gskKeyRegex = regexp.MustCompile(`\bgsk_[A-Za-z0-9]{40,}\b`)

func Groq(src string) (string, bool) {
	key := gskKeyRegex.FindString(src)
	if key == "" { // no match
		return "", false
	}
	return key, true
}

func init() {
	AllDetectors = append(AllDetectors, Groq)
}
