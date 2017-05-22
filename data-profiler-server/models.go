package main

import (
	"database/sql"
	"log"

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
}

// ModelOperator is the struct that represents an operator for a given dataset
type ModelOperator struct {
	ID          string
	Path        string
	Description string
}

// FUNCTIOND

// dbConnect is responsible to establish the connection with the DB backend.
// Written as separate function to increase modularity between the different
// DB backends.
func dbConnect() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", Conf.Database)
	return db, err
}

func modelDatasetsList() []*ModelDataset {
	db, err := dbConnect()
	if err != nil {
		log.Println(err)
	}
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
	db, err := dbConnect()
	if err != nil {
		log.Println(err)
		return nil
	}
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

func modelDatasetGetFiles(path string) *ModelDataset {
	// FIXME: add implementation
	return nil
}

func modelOperatorsGet(datasetId string) []*ModelOperator {
	// FIXME: add implementation
	return nil
}
