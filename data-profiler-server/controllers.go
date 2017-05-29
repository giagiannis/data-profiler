package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

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
		m.Matrices = modelSimilarityMatrixGetByDataset(m.ID)
		//	m.Estimators = modelDatasetGetEstimators(m.ID)
		m.Operators = modelOperatorGetByDataset(m.ID)
	}
	return m
}

// /download/
func controllerDownload(w http.ResponseWriter, r *http.Request) Model {
	fileType := r.URL.Query().Get("type")
	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	var filePath string
	if fileType == "datafile" {
		m := modelDatasetGetInfo(id)
		if m != nil {
			filePath = m.Path + "/" + name
		}
	} else if fileType == "sm" {
		m := modelSimilarityMatrixGet(id)
		if m != nil {
			filePath = m.Path
		}
	} else if fileType == "coord" {
		m := modelCoordinatesGet(id)
		if m != nil {
			filePath = m.Path
		}
	} else if fileType == "operator" {
		m := modelOperatorGet(id)
		if m != nil {
			filePath = m.Path
		}
	} else if fileType == "scores" {
		m := modelScoresGet(id)
		if m != nil {
			filePath = m.Path
		}
	}

	if filePath != "" {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Println(err)
		}

		w.Header().Set("Content-Disposition",
			"attachment; filename="+name)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Write(content)
	} else {
		w.WriteHeader(404)
	}

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
	ret := modelDatasetGetInfo(m.DatasetID)
	ret.Files = modelDatasetGetFiles(m.DatasetID)
	return ret
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

// /operator/<id>/delete
func controllerOperatorDelete(w http.ResponseWriter, r *http.Request) Model {
	datasetID := r.URL.Query().Get("datasetID")
	_, id, _ := parseURL(r.URL.Path)
	modelOperatorDelete(id)
	http.Redirect(w, r, "/datasets/"+datasetID, 307)
	return nil
}

// TASK BASED urls

// /datasets/<id>/newsm
func controllerDatasetNewSM(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	action := r.URL.Query().Get("action")
	if action != "submit" {
		m := modelDatasetGetInfo(id)
		scripts := make(map[string]string)
		for _, f := range Conf.Scripts.Analysis {
			scripts[f] = path.Base(f)
		}
		return struct {
			ID      string
			Scripts map[string]string
		}{ID: m.ID, Scripts: scripts}
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
	action := r.URL.Query().Get("action")
	_, id, _ := parseURL(r.URL.Path)
	if action != "submit" { // render the form
		return struct{ ID, DatasetID string }{id, datasetID}
	}
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	conf := map[string]string{"k": r.PostFormValue("k")}
	task := NewMDSComputationTask(id, datasetID, conf)
	TEngine.Submit(task)
	http.Redirect(w, r, "/tasks/", 307)
	return nil

}

// /operator/<id>/run
func controllerOperatorRun(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	task := NewOperatorRunTask(id)
	TEngine.Submit(task)
	http.Redirect(w, r, "/tasks/", 307)
	return nil
}

// /coords/<matrixid>
func controllerCoordsView(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	return modelCoordinatesGetByMatrix(id)
}

// /coords/<id>/visual
func controllerCoordsVisual(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	m := modelCoordinatesGet(id)
	if m == nil {
		log.Println("Coordinates file not found")
		return nil
	}
	cnt, err := ioutil.ReadFile(m.Path)
	if err != nil {
		log.Println(err)
	}

	sm := modelSimilarityMatrixGet(m.matrixID)
	if sm == nil {
		log.Println("SM not found")
		return nil
	}
	datasetID := sm.DatasetID
	fileNames := ""
	files := modelDatasetGetFiles(datasetID)
	for i, n := range files {
		fileNames += n
		if i < len(files)-1 {
			fileNames += "\n"
		}
	}
	operators := modelOperatorGetByDataset(datasetID)
	return struct {
		Coordinates, Labels string
		Operators           []*ModelOperator
	}{
		Coordinates: string(cnt),
		Labels:      fileNames,
		Operators:   operators}
}

// /dataset/<id>/newop
func controllerDatasetNewOP(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	action := r.URL.Query().Get("action")
	if action != "submit" { // render stuff for the form
		return modelDatasetGetInfo(id)
	}
	err := r.ParseMultipartForm(2 << 20)
	if err != nil {
		log.Println(err)
	}
	f, header, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
	}
	operatorDescription := r.PostForm["description"][0]
	operatorCnt, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
	}
	operatorFilename := header.Filename

	//	log.Println(operatorName, operatorDescription, len(operatorCnt), operatorFilename)
	modelOperatorInsert(id, operatorDescription, operatorFilename, operatorCnt)

	http.Redirect(w, r, "/datasets/"+id, 307)
	return nil
}

// /datasets/new/new
func controllerDatasetNew(w http.ResponseWriter, r *http.Request) Model {
	action := r.URL.Query().Get("action")
	if action != "submit" { // render stuff for the form
		return nil
	}
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	datasetName := r.PostForm["name"][0]
	datasetDescription := r.PostForm["description"][0]
	datasetPath := r.PostForm["path"][0]
	id := modelDatasetInsert(datasetName, datasetDescription, datasetPath)
	http.Redirect(w, r, "/datasets/"+id, 307)
	return nil
}

// /scores/<id>/text
func controllerScoresText(w http.ResponseWriter, r *http.Request) Model {
	_, id, _ := parseURL(r.URL.Path)
	m := modelScoresGet(id)
	if m == nil {
		return nil
	}
	cnt, err := ioutil.ReadFile(m.Path)
	if err != nil {
		log.Println(err)
	}
	s := core.NewDatasetScores()
	s.Deserialize(cnt)
	buffer := new(bytes.Buffer)
	for k, v := range s.Scores {
		buffer.WriteString(fmt.Sprintf("%s:%.5f\n", k, v))
	}
	w.Write(buffer.Bytes())
	return nil
}
