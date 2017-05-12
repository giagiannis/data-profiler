package main

import (
	"html/template"
	"net/http"
	"strings"
)

func viewHandler(w http.ResponseWriter, r *http.Request) {
	a := strings.Split(r.URL.Path, "/")
	page := "about.html"
	if len(a) > 1 && a[1] != "" {
		page = a[1] + ".html"
	}
	p := new(Page)
	p.Header = "DP"
	p.Title = "My cool title"
	p.Body = "Lorem Ipsum"

	t, err := template.ParseFiles(TEMPLATES_DIR+"/"+page, TEMPLATES_DIR+"/base.html")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		t, _ = template.ParseFiles(TEMPLATES_DIR+"/error.html", TEMPLATES_DIR+"/base.html")
	}
	t.Execute(w, p)
}
