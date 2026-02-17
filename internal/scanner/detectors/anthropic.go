// anthropic api keys
package detectors

import (
	"regexp"
)

var antKeyRegex = regexp.MustCompile(`sk-ant-[a-z0-9]{5,7}-[A-Za-z0-9_-]{90,110}`)

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
