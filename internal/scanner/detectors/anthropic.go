// anthropic api keys
package detectors

import (
	"regexp"
)

var antKeyRegex = regexp.MustCompile(`\bsk-ant-[A-Za-z0-9]{20,}\b`) // compile for xtra performance

func Anthropic(src string) (string, bool) {
	key := antKeyRegex.FindString(src)
	if key == "" { // no match
		return "", false
	}
	return key, true
}

func init() {
	AllDetectors = append(AllDetectors, Anthropic)
}
