package frontend

import (
	"embed"
)

//go:embed z2/*.js
//go:embed z2/css/*.css
//go:embed z2/css/purecss/*.css
//go:embed z2/css/notyf/*.css
//go:embed z2/pages/*.html
//go:embed z2/sounds/*.mp3
var FS embed.FS

//go:embed env_gin.js
var EnvWebJSTemplate []byte

//go:embed templates/*.html
var TemplatesFS embed.FS
