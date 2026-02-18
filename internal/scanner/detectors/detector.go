package detectors

import "strings"

type DetectorFunc func(src string) (string, bool, string)

var AllDetectors []DetectorFunc

func EnsureKeyIsntSpam(key string) bool {
	lower := strings.ToLower(key)
	if strings.Count(lower, "x") >= 6 {
		return false // spam
	}
	if strings.Contains(lower, "your_api_key") || strings.Contains(lower, "placeholder") {
		return false // spam
	}
	return true // not spam
}
