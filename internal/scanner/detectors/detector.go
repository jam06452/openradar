package detectors

type DetectorFunc func(src string) (string, bool, string)

var AllDetectors []DetectorFunc
