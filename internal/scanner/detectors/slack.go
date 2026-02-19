// Slack api keys
package detectors

import (
	"regexp"
)

var slackKeyRegex = regexp.MustCompile(`xox[bpaedcs]-[A-Za-z0-9-]{10,72}`)

func Slack(src string) (string, bool, string) {
	key := slackKeyRegex.FindString(src)
	if key == "" { // no match
		return "", false, "slack"
	}
	return key, true, "slack"
}

func init() {
	AllDetectors = append(AllDetectors, Slack)
}
