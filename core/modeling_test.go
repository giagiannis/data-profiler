package core

import "testing"

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
		t.Log(err)
		t.FailNow()
	}

	m := NewModeler(datasets, 0.2, coords, eval)

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
