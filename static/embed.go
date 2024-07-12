package static

import (
	"embed"
)

//go:embed *.html *.css sounds/*.mp3
//go:embed dist/bundle.js
//go:embed dist/synchronizer.js
var FS embed.FS
