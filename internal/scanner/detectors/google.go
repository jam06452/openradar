// Google api keys
package detectors

import (
	"regexp"
)

var googleAPIRegex = regexp.MustCompile(`\bAIzaSy[A-Za-z0-9-_]{35,}\b`)

func Google(src string) (string, bool) {
	key := googleAPIRegex.FindString(src)
	if key == "" { // no match
		return "", false
	}
	return key, true
}

func init() {
	AllDetectors = append(AllDetectors, Google)
}
