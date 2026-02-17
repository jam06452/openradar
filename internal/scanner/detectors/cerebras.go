// cerebras api keys
package detectors

import (
	"regexp"
)

var cskRegex = regexp.MustCompile(`csk-[A-Za-z0-9]{32,48}`)

func Cerebras(src string) (string, bool) {
	key := cskRegex.FindString(src)
	if key == "" { // no match
		return "", false
	}
	return key, true
}

func init() {
	AllDetectors = append(AllDetectors, Cerebras)
}
