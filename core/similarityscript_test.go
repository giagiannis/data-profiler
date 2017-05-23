package core

import (
	"os"
	"testing"
)

func TestScriptSimilarityDatasetAnalysis(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 20, 4)
	est := new(ScriptSimilarityEstimator)
	est.datasets = datasets
	conf := map[string]string{
		"concurrency": "10",
		"script":      analysisScript,
		"norm":        "1",
	}
	est.Configure(conf)
	results := est.analyzeDatasets()
	if len(results) != len(datasets) || len(results) == 0 {
		t.Log("Not all datasets analyzed")
		t.FailNow()
	}

	size := len(results[0])
	for _, val := range results {
		if len(val) != size {
			t.Log("Incorrect number of elements")
			t.FailNow()
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
func TestScriptSimilarityCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_SCRIPT, datasets)
	conf := map[string]string{
		"concurrency": "10",
		"script":      analysisScript,
		"norm":        "1",
	}
	est.Configure(conf)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.SimilarityMatrix()
	for i := range datasets {
		for j := range datasets {
			if s.Get(i, j) != s.Get(j, i) {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
		}
	}
	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestScriptSimilarityComputeAppxThres(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_SCRIPT, datasets)
	conf := map[string]string{
		"concurrency": "10",
		"script":      analysisScript,
		"norm":        "1",
	}
	est.Configure(conf)
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"threshold": 0.95,
		},
	}
	est.SetPopulationPolicy(pol)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.SimilarityMatrix()
	for i := range datasets {
		for j := range datasets {
			if s.Get(i, j) != s.Get(j, i) {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
		}
	}
	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestScriptSimilarityComputeAppxCnt(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_SCRIPT, datasets)
	conf := map[string]string{
		"concurrency": "10",
		"script":      analysisScript,
		"norm":        "1",
	}
	est.Configure(conf)
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"count": 10,
		},
	}
	est.SetPopulationPolicy(pol)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.SimilarityMatrix()

	for i := range datasets {
		for j := range datasets {
			if s.Get(i, j) != s.Get(j, i) {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
		}
	}
	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestScriptSimilaritySerialization(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 100, 4)
	//	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACCARD, datasets)
	est := *new(ScriptSimilarityEstimator)
	est.datasets = datasets
	est.simType = scriptSimilarityTypeEuclidean
	est.analysisScript = analysisScript
	est.concurrency = 10
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_FULL,
		Parameters: map[string]float64{},
	}
	est.SetPopulationPolicy(pol)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	bytes := est.Serialize()

	newEst := *new(ScriptSimilarityEstimator)
	newEst.Deserialize(bytes)
	if est.concurrency != newEst.concurrency {
		t.Log("Concurrency differs")
		t.Fail()
	}

	if est.analysisScript != newEst.analysisScript {
		t.Log("Analysis script differs")
		t.Fail()
	}

	if est.simType != newEst.simType {
		t.Log("norm degree differs")
		t.Fail()
	}

	for i := range est.datasets {
		if est.datasets[i].Path() != newEst.datasets[i].Path() {
			t.Log("Dataset names are different", est.datasets[i], newEst.datasets[i])
			t.Fail()
		}
	}

	for i := 0; i < est.similarities.Capacity(); i++ {
		for j := 0; j < est.similarities.Capacity(); j++ {
			if est.similarities.Get(i, j) != newEst.similarities.Get(i, j) {
				t.Log("SM differs", i, j)
				t.Fail()
			}
		}
	}

	for k, v := range est.inverseIndex {
		if newEst.inverseIndex[k] != v {
			t.Log("Inverse index failed")
			t.Fail()
		}
	}

	for i, arr := range est.datasetCoordinates {
		for j, v := range arr {
			if v != newEst.datasetCoordinates[i][j] {
				t.Log("Coordinates failed")
				t.Fail()
			}
		}
	}
	if newEst.Similarity(datasets[0], datasets[1]) != newEst.SimilarityMatrix().Get(0, 1) {
		t.Log("Something is seriously wrong here", newEst.SimilarityMatrix().Get(0, 1), newEst.Similarity(datasets[0], datasets[1]))
		t.Fail()
	}

}

func TestScriptSimilarityCosine(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 20, 3)
	s := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_SCRIPT, datasets)
	s.Configure(map[string]string{
		"script":      analysisScript,
		"concurrency": "1",
		"type":        "cosine",
	})
	err := s.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	mat := s.SimilarityMatrix()
	for i := range datasets {
		for j := range datasets {
			if mat.Get(i, j) < -1.0 || mat.Get(i, j) > 1.0 {
				t.Logf("Cosine Similarity out of bounds between [%d, %d] -> %.3f\n", i, j, mat.Get(i, j))
				//				t.FailNow()
			}
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
