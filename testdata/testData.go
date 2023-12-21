package testdata

import (
	"embed"
)

//go:embed *.tell
var Tell embed.FS

//go:embed *.json
var Json embed.FS
