package dist

import (
	"embed"
)

//go:embed *.js
//go:embed *.css
//go:embed purecss/*.css
//go:embed html/*.html
//go:embed sounds/*.mp3
var FS embed.FS
