package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

// /datasets/
func controllerDatasetList(w http.ResponseWriter, r *http.Request) Model {
	return modelDatasetsList()
}

// /datasets/<id>/
func controllerDatasetView(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	m := modelDatasetGetInfo(id)
	files := modelDatasetGetFiles(m.ID)
	m.Files = files
	m.Matrices = modelDatasetGetMatrices(m.ID)
	//	m.Estimators = modelDatasetGetEstimators(m.ID)
	return m
}

// /datasets/<id>/newsm
func controllerDatasetNewSM(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	action := r.URL.Query().Get("action")
	if action != "submit" {
		_, id, _ := parseURL(r.URL.Path)
		return modelDatasetGetInfo(id)
	}
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	conf := make(map[string]string)
	for k, v := range r.PostForm {
		if len(v) > 0 {
			conf[k] = v[0]
		} else {
			conf[k] = ""
		}
	}
	TEngine.Submit(NewSMComputationTask(id, conf))
	http.Redirect(w, r, "/tasks/", 301)
	return nil
}

// /download/
func controllerDownload(w http.ResponseWriter, r *http.Request) Model {
	fileType := r.URL.Query().Get("type")
	datasetID := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	var filePath string
	if fileType == "datafile" {
		filePath = modelDatasetGetInfo(datasetID).Path + "/" + name
	} else if fileType == "sm" {
		filePath = modelSimilarityMatrixGet(datasetID).Path
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Disposition",
		"attachment; filename="+name)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Write(content)

	return nil
}

// /tasks/
func controllerTasksList(w http.ResponseWriter, r *http.Request) Model {
	return TEngine.Tasks
}

// /sm/<id>/visual
func controllerSMVisual(w http.ResponseWriter, r *http.Request) Model {
	return nil
}
