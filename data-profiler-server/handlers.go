package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

type cntTmpltCouple struct {
	cnt func(http.ResponseWriter, *http.Request) Model
	tmp string
}

// TEMPLATE_DEPENDENCIES lists the necessary templates that need to be rendered
// for each template
var TEMPLATE_DEPENDENCIES = map[string][]string{
	"about.html":         []string{"base.html"},
	"datasets.html":      []string{"base.html"},
	"datasets_view.html": []string{"base.html"},
	"tasks.html":         []string{"base.html"},
	"error.html":         []string{"base.html"},
}

// ROUTING_CONTROLLER_TEMPLATES hold the controller and the respective template
// that need to be rendered for each possible path
var ROUTING_CONTROLLER_TEMPLATES = map[string]cntTmpltCouple{
	"datasets/":     cntTmpltCouple{controllerDatasetList, "datasets.html"},
	"datasets/view": cntTmpltCouple{controllerDatasetView, "datasets_view.html"},
	"about/":        cntTmpltCouple{nil, "about.html"},
	"tasks/":        cntTmpltCouple{nil, "tasks.html"},
}

func uiHandler(w http.ResponseWriter, r *http.Request) {
	cnt, t := selectControllerAndTemplate(r.URL.Path)
	var m Model
	if cnt != nil {
		m = cnt(w, r)
	}
	if t.Lookup("error.html") != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	t.Execute(w, m)
}

func selectControllerAndTemplate(url string) (func(http.ResponseWriter, *http.Request) Model, *template.Template) {
	model, id, cmd := parseURL(url)
	if id != "" && cmd == "" { // default action is view
		cmd = "view"
	}
	route := model + "/" + cmd

	var tmplt string
	var cnt func(http.ResponseWriter, *http.Request) Model

	if coup, ok := ROUTING_CONTROLLER_TEMPLATES[route]; ok {
		cnt = coup.cnt
		tmplt = coup.tmp
	}
	t := loadTemplate(tmplt)
	return cnt, t
}

// loadTemplate attempts to load the specified template, else returns the
// error page
func loadTemplate(templateName string) *template.Template {
	deps, ok := TEMPLATE_DEPENDENCIES[templateName]
	if !ok { // error
		deps = TEMPLATE_DEPENDENCIES["error.html"]
		templateName = "error.html"
	}
	newDeps := make([]string, len(deps))
	copy(newDeps, deps)
	deps = newDeps
	for i := range deps {
		deps[i] = Conf.Server.Dirs.Templates + "/" + deps[i]
	}
	deps = append(deps, Conf.Server.Dirs.Templates+"/"+templateName)
	t, err := template.ParseFiles(deps...)
	if err != nil {
		log.Println(err)
	}
	return t
}

func parseURL(url string) (string, string, string) {
	if url == "/" {
		url = "/datasets/"
	}
	arr := strings.Split(url, "/")[1:]
	if arr[len(arr)-1] == "" {
		arr = arr[0 : len(arr)-1]
	}
	model, id, cmd := arr[0], "", ""
	if len(arr) > 1 {
		id = arr[1]
	}
	if len(arr) > 2 {
		cmd = arr[2]
	}
	return model, id, cmd
}
