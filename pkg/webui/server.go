package webui

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"

	webuifs "github.com/christophwitzko/wg-hub/webui"
)

func getWebuiFS() http.FileSystem {
	s, err := fs.Sub(webuifs.FS, "out")
	if err != nil {
		panic(err)
	}
	return http.FS(s)
}

func getWebuiServer() http.HandlerFunc {
	wfs := getWebuiFS()
	fileServer := http.FileServer(wfs)
	return func(w http.ResponseWriter, r *http.Request) {
		rPath := path.Clean(r.URL.Path)
		if !strings.HasSuffix(rPath, "/") {
			_, err := wfs.Open(rPath)
			// if the file does not exist, try to serve the html file
			if errors.Is(err, fs.ErrNotExist) {
				rPath += ".html"
				r.URL.Path = rPath
			}
		}
		fileServer.ServeHTTP(w, r)
	}
}
