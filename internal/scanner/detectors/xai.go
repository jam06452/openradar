// xAI api keys
package detectors

import (
	"regexp"
)

var xaiKeyRegex = regexp.MustCompile(`\bxai-[A-Za-z0-9]{40,}\b`)

func xAI(src string) (string, bool) {
	key := xaiKeyRegex.FindString(src)
	if key == "" { // no match
		return "", false
	}
	return key, true
}

func init() {
	AllDetectors = append(AllDetectors, xAI)
}
