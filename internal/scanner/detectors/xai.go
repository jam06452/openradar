// xAI api keys
package detectors

import (
	"regexp"
)

var xaiKeyRegex = regexp.MustCompile(`xai-[A-Za-z0-9]{32,128}`)

func xAI(src string) (string, bool, string) {
	key := xaiKeyRegex.FindString(src)
	if key == "" { // no match
		return "", false, "xai"
	}
	return key, true, "xai"
}

func init() {
	AllDetectors = append(AllDetectors, xAI)
}
