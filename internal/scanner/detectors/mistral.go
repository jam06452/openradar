// Mistral api keys
package detectors

import (
	"regexp"
)

var misKeyRegex = regexp.MustCompile(`\bmis_[A-Za-z0-9]{40,}\b`)

func Mistral(src string) (string, bool) {
	key := misKeyRegex.FindString(src)
	if key == "" { // no match
		return "", false
	}
	return key, true
}

func init() {
	AllDetectors = append(AllDetectors, Mistral)
}
