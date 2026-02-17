// Google api keys
package detectors

import (
	"regexp"
)

var googleAPIRegex = regexp.MustCompile(`AIzaSy[A-Za-z0-9_-]{33,}`)

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
