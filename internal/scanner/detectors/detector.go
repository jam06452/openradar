package detectors

type DetectorFunc func(src string) (string, bool)

var AllDetectors []DetectorFunc
