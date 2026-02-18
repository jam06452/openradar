// openrouter api keys
package detectors

import (
	"regexp"
)

var openrouterRegex = regexp.MustCompile(`sk-or-v1-[a-f0-9]{64}`)

func OpenRouter(src string) (string, bool, string) {
	key := openrouterRegex.FindString(src)
	if key == "" {
		return "", false, "openrouter"
	}
	return key, true, "openrouter"
}

func init() {
	AllDetectors = append(AllDetectors, OpenRouter)
}
