package templates

import "embed"

//go:embed emails/*.html
var Files embed.FS
