// Mistral api keys
package detectors

import (
	"regexp"
)

var misKeyRegex = regexp.MustCompile(`mis(?:tral)?_[A-Za-z0-9]{32,56}`)

func Mistral(src string) (string, bool, string) {
	key := misKeyRegex.FindString(src)
	if key == "" { // no match
		return "", false, "mistral"
	}
	return key, true, "mistral"
}

func init() {
	AllDetectors = append(AllDetectors, Mistral)
}
