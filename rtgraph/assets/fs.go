package assets

import "embed"

//go:embed *.html *.js *.css
var FS embed.FS
