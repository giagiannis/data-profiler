package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

// Model is the interface returned by all the model functions
type Model interface{}

// ModelDataset is the struct that represents a set of datasets
type ModelDataset struct {
	Id          string
	Name        string
	Path        string
	Description string
}

func modelDatasetsList() []*ModelDataset {
	db, err := sql.Open("sqlite3", Conf.Database)
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
		rows.Scan(&obj.Id, &obj.Path, &obj.Name, &obj.Description)
		result = append(result, obj)
	}
	return result
}

func modelDatasetGet(id string) *ModelDataset {
	db, err := sql.Open("sqlite3", Conf.Database)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM datasets WHERE id == " + id)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	rows.Next()
	obj := new(ModelDataset)
	rows.Scan(&obj.Id, &obj.Path, &obj.Name, &obj.Description)
	return obj
}
