package main

import "net/http"

func controllerDatasetList(w http.ResponseWriter, r *http.Request) Model {
	return modelDatasetsList()
}
func controllerDatasetView(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	return modelDatasetGet(id)
}
