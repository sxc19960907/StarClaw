package daemon

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed webui/*
var webUIFS embed.FS

func (s *Server) handleWebRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/app/", http.StatusFound)
}

func (s *Server) handleWebAppRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/app/", http.StatusFound)
}

func (s *Server) handleWebApp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/app/" {
		http.NotFound(w, r)
		return
	}
	data, err := webUIFS.ReadFile("webui/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleWebAsset(w http.ResponseWriter, r *http.Request) {
	assets, err := fs.Sub(webUIFS, "webui/assets")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	http.StripPrefix("/app/assets/", http.FileServer(http.FS(assets))).ServeHTTP(w, r)
}
