package web

import (
	"embed"
	_ "embed"
)

// content holds our static web server content.
//
//go:embed *
var FS embed.FS
