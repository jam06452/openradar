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
	if strings.Contains(lower, "abcdefg") {
		return false // spam
	}
	if strings.Contains(lower, "12345678") {
		return false // spam
	}
	if strings.Contains(lower, "123") {
		return false //spam
	}
	if strings.Contains(lower, "abc") {
		return false
	}

	return true // not spam
}
