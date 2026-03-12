package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var staticFiles embed.FS

// StaticHandler returns an http.Handler that serves embedded static assets.
func StaticHandler() http.Handler {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic("web: static sub-FS: " + err.Error())
	}
	return http.FileServer(http.FS(sub))
}
