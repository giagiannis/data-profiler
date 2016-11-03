package core

import (
	"os"
	"testing"
)

func TestDatatesetIntersection(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 2, 10)
	res := DatasetsIntersection(&datasets[0], &datasets[1])
	if len(datasets[0].Data()) < len(res) ||
		len(datasets[1].Data()) < len(res) {
		t.Log(len(datasets[0].Data()))
		t.Log(len(datasets[1].Data()))
		t.FailNow()
	}

	for _, d := range datasets {
		os.Remove(d.Path())
	}
}

func TestDatasetUnion(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 2, 10)
	res := DatasetsUnion(&datasets[0], &datasets[1])
	t.Log(len(res))
	if len(datasets[0].Data()) > len(res) ||
		len(datasets[1].Data()) > len(res) {
		t.Log(len(datasets[0].Data()))
		t.Log(len(datasets[1].Data()))
		t.FailNow()
	}

	for _, d := range datasets {
		os.Remove(d.Path())
	}
}
