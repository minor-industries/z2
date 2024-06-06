package static

import (
	"embed"
)

//go:embed *.html *.js *.css sounds/*.mp3
var FS embed.FS
