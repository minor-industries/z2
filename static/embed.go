package static

import (
	"embed"
)

//go:embed *.html *.css sounds/*.mp3
//go:embed dist/z2-bundle.js
var FS embed.FS
