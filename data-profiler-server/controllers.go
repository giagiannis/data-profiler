package main

import "net/http"

func controllerDatasetList(w http.ResponseWriter, r *http.Request) Model {
	return modelDatasetsList()
}
