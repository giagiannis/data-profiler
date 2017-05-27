package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/giagiannis/data-profiler/core"
	_ "github.com/mattn/go-sqlite3"
)

// Models

// Model is the interface returned by all the model functions
type Model interface{}

// ModelDataset is the struct that represents a set of datasets
type ModelDataset struct {
	ID          string
	Name        string
	Path        string
	Description string
	Files       []string
	Operators   []*ModelOperator
	Matrices    []*ModelSimilarityMatrix
	Estimators  []*ModelEstimator
}

// ModelOperator is the struct that represents an operator for a given dataset
type ModelOperator struct {
	ID          string
	Path        string
	Description string
	Name        string
	Filename    string
	DatasetID   string
}

// ModelEstimator represents an estimator object
type ModelEstimator struct {
	ID            string
	Path          string
	Filename      string
	Configuration map[string]string
	DatasetID     string
}

// ModelSimilarityMatrix represents a similarity matrix
type ModelSimilarityMatrix struct {
	ID            string
	Path          string
	Filename      string
	Configuration map[string]string
	DatasetID     string
}

// ModelCoordinates represents a set of coordinates
type ModelCoordinates struct {
	ID       string
	Path     string
	Filename string
	matrixID string
	K        string
	GOF      string
}

// FUNCTIONS

// dbConnect is responsible to establish the connection with the DB backend.
// Written as separate function to increase modularity between the different
// DB backends.
func dbConnect() *sql.DB {
	db, err := sql.Open("sqlite3", Conf.Database)
	if err != nil {
		log.Println(err)
	}
	return db
}

func modelDatasetsList() []*ModelDataset {
	db := dbConnect()
	defer db.Close()
	rows, err := db.Query("SELECT * FROM datasets")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	result := make([]*ModelDataset, 0)
	for rows.Next() {
		obj := new(ModelDataset)
		rows.Scan(&obj.ID, &obj.Path, &obj.Name, &obj.Description)
		result = append(result, obj)
	}
	return result
}

func modelDatasetGetInfo(id string) *ModelDataset {
	db := dbConnect()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM datasets WHERE id == " + id)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	if rows.Next() {
		obj := new(ModelDataset)
		rows.Scan(&obj.ID, &obj.Path, &obj.Name, &obj.Description)
		return obj
	}
	return nil
}

func modelDatasetGetFiles(id string) []string {
	var results []string
	m := modelDatasetGetInfo(id)
	if m == nil {
		return nil
	}
	path := modelDatasetGetInfo(id).Path
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err)
	}
	for _, f := range fs {
		if !f.IsDir() {
			results = append(results, f.Name())
		}
	}
	return results
}

func modelSimilarityMatrixGetByDataset(id string) []*ModelSimilarityMatrix {
	db := dbConnect()
	defer db.Close()
	var results []*ModelSimilarityMatrix
	rows, err := db.Query("SELECT id, path, filename,configuration " +
		" FROM matrices WHERE datasetid == " + id)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		obj := new(ModelSimilarityMatrix)
		confString := ""
		rows.Scan(&obj.ID, &obj.Path, &obj.Filename, &confString)
		conf := make(map[string]string)
		err := json.Unmarshal([]byte(confString), &conf)
		if err != nil {
			log.Println(err)
		}
		obj.Configuration = conf
		results = append(results, obj)
	}
	return results
}

func modelEstimatorGetByDataset(id string) []*ModelEstimator {
	db := dbConnect()
	defer db.Close()
	var results []*ModelEstimator
	rows, err := db.Query("SELECT id, path, filename,configuration " +
		" FROM estimators WHERE datasetid == " + id)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		obj := new(ModelEstimator)
		confString := ""
		rows.Scan(&obj.ID, &obj.Path, &obj.Filename, &confString)
		conf := make(map[string]string)
		err := json.Unmarshal([]byte(confString), &conf)
		if err != nil {
			log.Println(err)
		}
		obj.Configuration = conf
		results = append(results, obj)
	}
	return results
}

// modelSimilarityMatrixInsert inserts a new SM and returns the newly created Id
func modelSimilarityMatrixInsert(datasetID string, buffer []byte, conf map[string]string) string {
	dts := modelDatasetGetInfo(datasetID)
	filePath := writeBufferToFile(dts, "matrices", buffer)
	db := dbConnect()
	defer db.Close()
	stmt, err := db.Prepare(
		"INSERT INTO matrices(path,filename,configuration,datasetid) " +
			"VALUES(?,?,?,?)")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
	}
	confString, err := json.Marshal(conf)
	if err != nil {
		log.Println(err)
	}
	res, err := stmt.Exec(filePath,
		path.Base(filePath),
		confString,
		dts.ID)
	if err != nil {
		log.Println(err)
	}
	resultInt, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
	}
	return fmt.Sprintf("%d", resultInt)
}

func modelSimilarityMatrixGet(id string) *ModelSimilarityMatrix {
	db := dbConnect()
	defer db.Close()

	rows, err := db.Query("SELECT id,path,filename,configuration,datasetid" +
		" FROM matrices WHERE id == " + id)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	if rows.Next() {
		obj := new(ModelSimilarityMatrix)
		confString := ""
		rows.Scan(
			&obj.ID,
			&obj.Path,
			&obj.Filename,
			&confString,
			&obj.DatasetID)
		obj.Configuration = make(map[string]string)
		json.Unmarshal([]byte(confString), &obj.Configuration)
		return obj
	}
	return nil
}

func modelSimilarityMatrixDelete(id string) *ModelSimilarityMatrix {
	m := modelSimilarityMatrixGet(id)
	if m != nil {
		os.Remove(m.Path)
	}
	db := dbConnect()
	defer db.Close()
	stmt, err := db.Prepare("DELETE FROM matrices WHERE id == ?")
	if err != nil {
		log.Println(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func modelEstimatorInsert(datasetID, matrixID string, buffer []byte, conf map[string]string) {
	dts := modelDatasetGetInfo(datasetID)
	filePath := writeBufferToFile(dts, "estimators", buffer)
	db := dbConnect()
	defer db.Close()
	stmt, err := db.Prepare(
		"INSERT INTO estimators(path,filename,configuration,datasetid,matrixid) " +
			"VALUES(?,?,?,?,?)")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
	}
	confString, err := json.Marshal(conf)
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(filePath,
		path.Base(filePath),
		confString,
		dts.ID,
		matrixID)
	if err != nil {
		log.Println(err)
	}
}

func modelEstimatorGet(id string) *ModelEstimator {
	db := dbConnect()
	defer db.Close()

	rows, err := db.Query("SELECT id,path,filename,configuration" +
		" FROM estimator WHERE id == " + id)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	if rows.Next() {
		obj := new(ModelEstimator)
		rows.Scan(&obj.ID, &obj.Path, &obj.Filename,
			&obj.Configuration)
		return obj
	}
	return nil
}

func modelEstimatorGetBySM(id string) *ModelEstimator {
	db := dbConnect()
	defer db.Close()

	rows, err := db.Query("SELECT id,path,filename,configuration" +
		" FROM estimator WHERE matrixid == " + id)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	if rows.Next() {
		obj := new(ModelEstimator)
		rows.Scan(&obj.ID, &obj.Path, &obj.Filename,
			&obj.Configuration)
		return obj
	}
	return nil
}
func modelCoordinatesGet(id string) *ModelCoordinates {
	db := dbConnect()
	defer db.Close()

	rows, err := db.Query("SELECT *" +
		" FROM coordinates WHERE id == " + id)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	if rows.Next() {
		obj := new(ModelCoordinates)
		rows.Scan(&obj.ID, &obj.Path, &obj.Filename, &obj.K,
			&obj.GOF, &obj.matrixID)
		return obj
	}
	return nil
}

func modelCoordinatesGetByMatrix(matrixId string) []*ModelCoordinates {
	db := dbConnect()
	defer db.Close()

	rows, err := db.Query("SELECT *" +
		" FROM coordinates WHERE matrixid == " + matrixId)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	var result []*ModelCoordinates
	for rows.Next() {
		obj := new(ModelCoordinates)
		rows.Scan(&obj.ID, &obj.Path, &obj.Filename, &obj.K,
			&obj.GOF, &obj.matrixID)
		result = append(result, obj)
	}
	return result
}

func modelCoordinatesInsert(coordinates []core.DatasetCoordinates, datasetID, K, GOF, matrixID string) {
	dts := modelDatasetGetInfo(datasetID)
	var coords [][]float64
	for _, c := range coordinates {
		coords = append(coords, c)
	}
	buffer := serializeCSVFile(coords)
	filePath := writeBufferToFile(dts, "coords", buffer)

	db := dbConnect()
	defer db.Close()
	stmt, err := db.Prepare(
		"INSERT INTO coordinates(path,filename,k,gof,matrixid) " +
			"VALUES(?,?,?,?,?)")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(filePath,
		path.Base(filePath),
		K,
		GOF,
		matrixID)
	if err != nil {
		log.Println(err)
	}
}

func modelOperatorInsert(datasetID, name, description, filename string, content []byte) {
	dts := modelDatasetGetInfo(datasetID)
	filePath := writeBufferToFile(dts, "operators", content)
	db := dbConnect()
	defer db.Close()
	stmt, err := db.Prepare(
		"INSERT INTO operators(name,description,path,filename,datasetid) " +
			"VALUES(?,?,?,?,?)")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(name,
		description,
		filePath,
		filename,
		datasetID)
	if err != nil {
		log.Println(err)
	}
}

func modelOperatorGet(id string) *ModelOperator {
	db := dbConnect()
	defer db.Close()

	rows, err := db.Query("SELECT *" +
		" FROM operators WHERE id == " + id)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	if rows.Next() {
		obj := new(ModelOperator)
		rows.Scan(&obj.ID, &obj.Name, &obj.Description,
			&obj.Path, &obj.Filename, &obj.DatasetID)
		return obj
	}
	return nil
}

func modelOperatorGetByDataset(id string) []*ModelOperator {
	db := dbConnect()
	defer db.Close()
	var results []*ModelOperator
	rows, err := db.Query("SELECT *" +
		" FROM operators WHERE datasetid == " + id)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		obj := new(ModelOperator)
		rows.Scan(&obj.ID, &obj.Name, &obj.Description,
			&obj.Path, &obj.Filename, &obj.DatasetID)
		results = append(results, obj)
	}
	return results
}

// utility functions
func writeBufferToFile(dts *ModelDataset, prefix string, buffer []byte) string {
	dstDir := dts.Path + "/" + prefix
	err := os.Mkdir(dstDir, 0777)
	if err != nil {
		log.Println(err)
	}
	dstPath := dstDir + "/" + prefix + currentTimeSuffix()
	err = ioutil.WriteFile(dstPath, buffer, 0777)
	if err != nil {
		log.Println(err)
	}
	return dstPath
}

func currentTimeSuffix() string {
	t := time.Now()
	y, m, d := t.Year(), int(t.Month()), t.Day()
	h, min, sec := t.Hour(), t.Minute(), t.Second()
	return fmt.Sprintf("%d%d%d%d%d%d", y, m, d, h, min, sec)
}

func serializeCSVFile(coords [][]float64) []byte {
	buffer := new(bytes.Buffer)
	for i := range coords {
		for j := range coords[i] {
			buffer.WriteString(fmt.Sprintf("%.5f", coords[i][j]))
			if j < len(coords[i])-1 {
				buffer.WriteString(",")
			}
		}
		if i < len(coords)-1 {
			buffer.WriteString("\n")
		}
	}
	return buffer.Bytes()
}

func readCSVFile(buffer []byte) [][]float64 {
	fileCont := string(buffer)
	lines := strings.Split(fileCont, "\n")
	var result [][]float64
	for _, line := range lines {
		var tuple []float64
		vals := strings.Split(line, ",")
		for _, v := range vals {
			floatVal, err := strconv.ParseFloat(v, 32)
			if err != nil {
				log.Println(err)
			}
			tuple = append(tuple, floatVal)
		}
		result = append(result, tuple)
	}
	return result
}
