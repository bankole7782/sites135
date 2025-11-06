package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

func serveWDl(projectName, port string) {
	rootPath, _ := GetRootPath()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		toFindPath := filepath.Join(rootPath, projectName, r.URL.Path)
		if r.URL.Path == "/" {
			toFindPath = filepath.Join(toFindPath, "index")
		}
		if !strings.Contains(toFindPath, ".") {
			toFindPath += ".html"
		}
		fmt.Println(toFindPath)

		http.ServeFile(w, r, toFindPath)
	})

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
