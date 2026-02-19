package detectors

import (
	"regexp"
)

var discordKeyRegex = regexp.MustCompile(`[MN][A-Za-z0-9]{23}\.[A-Za-z0-9_-]{6}\.[A-Za-z0-9_-]{27}`)

func Discord(src string) (string, bool, string) {
	key := discordKeyRegex.FindString(src)
	if key == "" {
		return "", false, "discord"
	}
	return key, true, "discord"
}

func init() {
	AllDetectors = append(AllDetectors, Discord)
}
