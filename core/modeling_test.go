package core

import (
	"math"
	"math/rand"
	"os"
	"testing"
)

func TestScriptBasedModelerConfiguration(t *testing.T) {
	m := new(ScriptBasedModeler)
	err := m.Configure(map[string]string{"foo": "bar"})
	if err == nil {
		t.Log("Should have returned an error!")
		t.FailNow()
	}
	err = m.Configure(map[string]string{"script": mlScript})
	if err != nil {
		t.Log("Should have not returned an error!")
		t.FailNow()
	}
}

func TestScriptBasedModelerRun(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 3)
	m, err := createModeler(datasets)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	m.Configure(map[string]string{"script": mlScriptAppx})
	err = m.Run()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if m.Samples() == nil || len(m.Samples()) == 0 {
		t.Log("Should have created samples")
		t.Fail()
	}
	if len(m.AppxValues()) != len(m.Datasets()) {
		t.Log("Wrong number of approximated values")
		t.Fail()
	}

	cleanDatasets(datasets)

}

func TestErrorMetrics(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 3)
	m, err := createModeler(datasets)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	m.Configure(map[string]string{"script": mlScriptAppx})
	if m.ErrorMetrics() != nil {
		t.Log("ErrorMetrics should have been nil")
		t.Fail()
	}
	err = m.Run()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	metrics := m.ErrorMetrics()
	if metrics["MSE-all"] < 0 || metrics["R^2-all"] > 1.0 || metrics["R^2-all"] < 0.0 {
		t.Log("Wrong metric values")
		t.Fail()
	}
	cleanDatasets(datasets)
}

func TestScriptBasedModelerMissingDatasets(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 3)
	noDeletedDatasets := 47
	m, err := createModeler(datasets)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	permutation := rand.Perm(len(datasets))
	for i := 0; i < noDeletedDatasets; i++ {
		path := datasets[permutation[i]].Path()
		os.Remove(path)
	}

	m.Configure(map[string]string{"script": mlScriptAppx})
	err = m.Run()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if m.Samples() == nil || len(m.Samples()) == 0 {
		t.Log("Should have created samples")
		t.Fail()
	}
	if len(m.AppxValues()) != len(m.Datasets()) {
		t.Log("Wrong number of approximated values")
		t.Fail()
	}
	if len(m.Samples()) != (50 - noDeletedDatasets) {
		t.Log("Expected", noDeletedDatasets, "samples and got", len(m.Samples()))
		t.Fail()
	}

	cleanDatasets(datasets)
}

func createModeler(datasets []*Dataset) (Modeler, error) {
	e := NewDatasetSimilarityEstimator(SimilarityTypeBhattacharyya, datasets)
	e.Configure(map[string]string{"partitions": "256"})
	e.Compute()
	sm := e.SimilarityMatrix()

	mds := NewMDScaling(sm, 3, mdsScript)
	mds.Compute()
	coords := mds.Coordinates()

	eval, err := NewDatasetEvaluator(OnlineEval, map[string]string{
		"script":  operatorScript,
		"testset": "",
	})
	if err != nil {
		return nil, err
	}

	m := NewModeler(datasets, 0.2, coords, eval)
	return m, nil
}

func TestErrorMetricFunctions(t *testing.T) {
	sizeA, sizeB := 100, 80
	arrayA, arrayB := make([]float64, sizeA), make([]float64, sizeB)
	for i := range arrayA {
		arrayA[i] = rand.Float64()
	}
	for i := range arrayB {
		arrayB[i] = rand.Float64()
	}
	mse, mape, rsq := MeanSquaredError(arrayA, arrayB),
		MeanAbsolutePercentageError(arrayA, arrayB),
		RSquared(arrayA, arrayB)
	if !math.IsNaN(mse) || !math.IsNaN(mape) || !math.IsNaN(rsq) {
		t.Log("MSE/MAPE/R^2 should have been NaN")
		t.Fail()
	}
	mse, mape, rsq = MeanSquaredError(arrayA, arrayA),
		MeanAbsolutePercentageError(arrayA, arrayA),
		RSquared(arrayA, arrayA)
	if mse != 0.0 || mape != 0.0 || rsq != 1.0 {
		t.Log("MSE/MAPE/R^2 should have been 0/0/1")
		t.Fail()
	}
	perc0, perc25, perc50,
		perc75, perc100 := Percentile(arrayA, 0),
		Percentile(arrayA, 25),
		Percentile(arrayA, 50),
		Percentile(arrayA, 75),
		Percentile(arrayA, 100)
	if perc0 > perc25 || perc25 > perc50 || perc50 > perc75 || perc75 > perc100 {
		t.Log("Percentiles gone wrong")
		t.Fail()
	}
}
