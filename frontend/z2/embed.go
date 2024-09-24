package z2

import (
	"embed"
)

//go:embed *.js
//go:embed css/*.css
//go:embed css/purecss/*.css
//go:embed html/*.html
//go:embed sounds/*.mp3
var FS embed.FS
