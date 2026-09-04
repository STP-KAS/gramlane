package web

import "embed"

//go:embed templates/*.html
var Templates embed.FS

//go:embed static/css/*.css static/js/*.js static/img/* static/manifest.webmanifest
var Static embed.FS
