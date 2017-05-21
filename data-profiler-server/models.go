package main

// Model is the interface returned by all the model functions
type Model interface{}

// ModelDataset is the struct that represents a set of datasets
type ModelDataset struct {
	Name        string
	Path        string
	Description string
}

func modelDatasetsList() map[string]*ModelDataset {
	DatasetModels := map[string]*ModelDataset{
		"1": &ModelDataset{Name: "Dataset 1", Path: "/opt/dat1", Description: "Used for meteorology"},
	}

	return DatasetModels
}

func modelDatasetGet(id string) *ModelDataset {
	// TODO: to be implemented
	return nil
}
