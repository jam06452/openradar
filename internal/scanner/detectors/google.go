// Google api keys
package detectors

import (
	"regexp"
)

var googleAPIRegex = regexp.MustCompile(`AIzaSy[A-Za-z0-9_-]{33,}`)

func Google(src string) (string, bool, string) {
	key := googleAPIRegex.FindString(src)
	if key == "" { // no match
		return "", false, "google"
	}
	return key, true, "google"
}

func init() {
	AllDetectors = append(AllDetectors, Google)
}
