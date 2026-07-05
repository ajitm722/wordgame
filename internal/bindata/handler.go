package bindata

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// FrontendHandler returns an HTTP handler that serves frontend files
// using the go-bindata generated Asset() function.
//
// In production mode (go-bindata without -debug), Asset() returns file
// bytes directly from memory — no disk I/O, single binary.
//
// In development mode (go-bindata -debug), Asset() reads the real file
// from disk. Webpack --watch can rebuild and write new files, and the
// next HTTP request serves the latest version — no recompilation needed.
//
// The handler implements the SPA catch-all:
//   - /assets/*   → serve the matching file (JS, CSS, images, etc.)
//   - everything else → serve index.html (React Router handles routing)
func FrontendHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		var assetPath string
		if strings.HasPrefix(path, "/assets/") {
			assetPath = path[1:] // strip leading "/" → "assets/bundle.js"
		} else {
			assetPath = "assets/index.html"
		}

		data, err := Asset(assetPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		contentType := mime.TypeByExtension(filepath.Ext(assetPath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)
		// allow the browser to cache assets for 1 year (content-hashed filenames = immutable)
		if strings.HasPrefix(path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		w.Write(data)
	})
}
