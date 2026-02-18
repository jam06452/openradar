package app

import "embed"

//go:embed dist public
var Dist embed.FS
