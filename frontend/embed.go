package frontend

import (
	"embed"
)

//go:embed z2/*.js
//go:embed z2/css/*.css
//go:embed z2/css/purecss/*.css
//go:embed z2/html/*.html
//go:embed z2/sounds/*.mp3
var FS embed.FS
