package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
	if m != nil {
		files := modelDatasetGetFiles(m.ID)
		m.Files = files
		m.Matrices = modelDatasetGetMatrices(m.ID)
		//	m.Estimators = modelDatasetGetEstimators(m.ID)
	}
	return m
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
	_, id, _ := parseURL(r.URL.Path)
	m := modelSimilarityMatrixGet(id)
	return modelDatasetGetInfo(m.DatasetID)
}

// /sm/<id>/csv/
func controllerSMtoCSV(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	m := modelSimilarityMatrixGet(id)
	sm := new(core.DatasetSimilarityMatrix)
	cnt, err := ioutil.ReadFile(m.Path)
	if err != nil {
		log.Println(err)
	}
	sm.Deserialize(cnt)
	w.Write([]byte("x,y,value\n"))
	files := modelDatasetGetFiles(m.DatasetID)
	for i := 0; i < sm.Capacity(); i++ {
		for j := 0; j < sm.Capacity(); j++ {
			w.Write([]byte(fmt.Sprintf("%s,%s,%.5f\n", files[i], files[j], sm.Get(i, j))))
		}
	}
	return nil
}

// /sm/<id>/delete
func controllerSMDelete(w http.ResponseWriter, r *http.Request) Model {
	datasetID := r.URL.Query().Get("datasetID")
	_, id, _ := parseURL(r.URL.Path)
	modelSimilarityMatrixDelete(id)
	http.Redirect(w, r, "/datasets/"+datasetID, 307)
	return nil
}

// TASK BASED urls

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
	http.Redirect(w, r, "/tasks/", 307)
	return nil
}

// /mds/<id>/run
func controllerMDSRun(w http.ResponseWriter, r *http.Request) Model {
	datasetID := r.URL.Query().Get("datasetID")
	_, id, _ := parseURL(r.URL.Path)
	// FIXME:correct that!
	conf := map[string]string{"k": "2"}
	task := NewMDSComputationTask(id, datasetID, conf)
	TEngine.Submit(task)
	http.Redirect(w, r, "/tasks/", 307)
	return nil
}
