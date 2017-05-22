package main

import (
	"errors"
	"net/http"
)

// /datasets/
func controllerDatasetList(w http.ResponseWriter, r *http.Request) (Model, error) {
	return modelDatasetsList(), nil
}

// /datasets/<id>/
func controllerDatasetView(w http.ResponseWriter, r *http.Request) (Model, error) {
	_, id, _ := parseURL(r.URL.Path)
	m := modelDatasetGetInfo(id)
	if m == nil {
		return nil, errors.New("Not found")
	}
	return m, nil
}
