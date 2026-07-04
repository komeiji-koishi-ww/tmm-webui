package app

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

// Embedding the static directory by name includes nested vendor assets such as
// Element Plus. The previous static/* pattern only embedded one level.
//
//go:embed static
var staticFS embed.FS

func staticHandler() http.Handler {
	devStatic := strings.TrimSpace(os.Getenv("TMMWEB_DEV_STATIC"))
	if devStatic != "" && devStatic != "0" && !strings.EqualFold(devStatic, "false") {
		staticDir := strings.TrimSpace(os.Getenv("TMMWEB_STATIC_DIR"))
		if staticDir == "" {
			staticDir = "internal/app/static"
		}
		return staticFileHandler(os.DirFS(staticDir), true)
	}
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return staticFileHandler(sub, true)
}

func staticFileHandler(fileSystem fs.FS, noStore bool) http.Handler {
	fileServer := http.FileServer(http.FS(fileSystem))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if noStore {
			w.Header().Set("Cache-Control", "no-store, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		if _, err := fs.Stat(fileSystem, path); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
