package main

import "net/http"

type Page struct {
	Title  string
	Header string
	Body   string
}

const (
	TEMPLATES_DIR = "data-profiler-server/templates"
	STATIC_DIR    = "data-profiler-server/static"
)

func main() {
	fs := http.FileServer(http.Dir(STATIC_DIR))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe("0.0.0.0:8080", nil)
}
