package revealgo

import (
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

//go:embed assets/revealjs
var revealjs embed.FS

//go:embed assets/templates/slide.html
var slideTemplate string

type Server struct {
	port int
}

type ServerParam struct {
	Path          string
	Theme         string
	OriginalTheme bool
	Transition    string
}

func (server *Server) Serve(param ServerParam) {
	port := 3000
	if server.port > 0 {
		port = server.port
	}
	fmt.Printf("accepting connections at http://*:%d/\n", port)
	http.Handle("/", &rootHandler{param: param})
	http.Handle("/revealjs/", &assetHandler{assetPath: "assets"})
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

type assetHandler struct {
	assetPath string
}

func (h *assetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filepath := h.assetPath + r.URL.Path
	data, err := revealjs.ReadFile(filepath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w = setResponse(w, filepath, data)
}

type rootHandler struct {
	param ServerParam
}

func (h *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	path, err := filepath.Rel("./", "."+urlPath)
	if err != nil {
		log.Fatalf("error:%v", err)
	}
	_, err = os.Stat(path)
	if err == nil {
		data, err := ioutil.ReadFile(path)
		if err == nil {
			w = setResponse(w, path, data)
			return
		}
	}

	tmpl := template.New("slide template")
	tmpl.Parse(slideTemplate)
	if err != nil {
		log.Printf("error:%v", err)
		http.NotFound(w, r)
		return
	}
	err = tmpl.Execute(w, h.param)
	if err != nil {
		log.Fatalf("error:%v", err)
	}
}

func detectContentType(path string, data []byte) string {
	switch {
	case strings.HasSuffix(path, ".css"):
		return "text/css"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	}
	return http.DetectContentType(data)
}

func setResponse(w http.ResponseWriter, path string, data []byte) http.ResponseWriter {
	mimeType := detectContentType(path, data)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	if _, err := w.Write(data); err != nil {
		log.Fatal("unable to write data.")
	}
	return w
}
