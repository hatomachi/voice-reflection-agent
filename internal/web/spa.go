package web

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type SPAHandler struct {
	staticFS http.FileSystem
}

func NewSPAHandler(fileSystem fs.FS) *SPAHandler {
	return &SPAHandler{
		staticFS: http.FS(fileSystem),
	}
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// APIリクエストはここには来ないが、念のため
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	path := filepath.Clean(r.URL.Path)
	if path == "/" || path == "." {
		path = "index.html"
	} else {
		path = strings.TrimPrefix(path, "/")
	}

	// 静的ファイルが存在するかチェック
	f, err := h.staticFS.Open(path)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") {
			// SPA fallback
			indexFile, err := h.staticFS.Open("index.html")
			if err != nil {
				http.Error(w, "index.html not found in embedded filesystem", http.StatusInternalServerError)
				return
			}
			defer indexFile.Close()

			stat, err := indexFile.Stat()
			if err != nil {
				http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
				return
			}
			rs, ok := indexFile.(io.ReadSeeker)
			if !ok {
				http.Error(w, "Failed to read index.html as ReadSeeker", http.StatusInternalServerError)
				return
			}
			http.ServeContent(w, r, "index.html", stat.ModTime(), rs)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if stat.IsDir() {
		// ディレクトリなら index.html
		h.ServeHTTP(w, r)
		return
	}

	http.FileServer(h.staticFS).ServeHTTP(w, r)
}
