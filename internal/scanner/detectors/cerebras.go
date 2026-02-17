// cerebras api keys
package detectors

import (
	"regexp"
)

var cskRegex = regexp.MustCompile(`csk-[A-Za-z0-9]{32,48}`)

func Cerebras(src string) (string, bool, string) {
	key := cskRegex.FindString(src)
	if key == "" { // no match
		return "", false, "cerebras"
	}
	return key, true, "cerebras"
}

func init() {
	AllDetectors = append(AllDetectors, Cerebras)
}
