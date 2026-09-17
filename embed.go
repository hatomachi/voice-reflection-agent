package main

import (
	"embed"
	"io/fs"
)

//go:embed all:web
var embeddedWebFS embed.FS

func getStaticFS() (fs.FS, error) {
	return fs.Sub(embeddedWebFS, "web")
}
