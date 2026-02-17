// xAI api keys
package detectors

import (
	"regexp"
)

var xaiKeyRegex = regexp.MustCompile(`xai-[A-Za-z0-9]{32,64}`)

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
