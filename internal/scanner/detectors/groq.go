// Groq api keys
package detectors

import (
	"regexp"
)

var gskKeyRegex = regexp.MustCompile(`gsk_[A-Za-z0-9]{32,56}`)

func Groq(src string) (string, bool, string) {
	key := gskKeyRegex.FindString(src)
	if key == "" { // no match
		return "", false, "groq"
	}
	return key, true, "groq"
}

func init() {
	AllDetectors = append(AllDetectors, Groq)
}
