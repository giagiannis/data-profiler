package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/giagiannis/data-profiler/core"
)

// /datasets/
func controllerDatasetList(w http.ResponseWriter, r *http.Request) Model {
	return modelDatasetsList()
}

// /datasets/<id>/
func controllerDatasetView(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	m := modelDatasetGetInfo(id)
	files := modelDatasetGetFiles(m.Path)
	m.Files = files
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
	log.Println(r.PostForm)
	//confParams := r.PostForm
	simType := r.PostFormValue("type")
	estType := core.NewDatasetSimilarityEstimatorType(simType)
	log.Println(estType)
	// TODO: do the rest of the stuff here
	myFunc := func() error {
		time.Sleep(10 * time.Second)
		return nil
	}
	task := new(Task)
	task.fnc = myFunc
	TEngine.Submit(task)
	http.Redirect(w, r, "/datasets/"+id, 301)
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

func controllerTasksList(w http.ResponseWriter, r *http.Request) Model {
	return TEngine.Tasks
}
