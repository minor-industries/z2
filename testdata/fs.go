package testdata

import "embed"

//go:embed *.txt
var FS embed.FS
