package core

import (
	"math"
	"math/rand"
	"os"
	"testing"
)

func TestCorrelationEstimatorCompute(t *testing.T) {
	datasets := createPoolBasedDatasetsStrict(1000, 500, 20, 4)
	est := new(CorrelationEstimator)
	est.datasets = datasets
	types := []string{"Pearson", "kendall", "spearman"}
	typeSelected := types[rand.Int()%3]
	conf := map[string]string{
		"concurrency": "10",
		"type":        typeSelected,
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

func TestCorrelationComputeAppxCnt(t *testing.T) {
	datasets := createPoolBasedDatasetsStrict(1000, 500, 20, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_CORRELATION, datasets)
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"count": 20,
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

func TestCorrelationComputeAppxThres(t *testing.T) {
	datasets := createPoolBasedDatasetsStrict(1000, 500, 20, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_CORRELATION, datasets)
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"threshold": 0.3,
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

func TestCorrelationEstimatorSerialization(t *testing.T) {
	datasets := createPoolBasedDatasetsStrict(1000, 500, 20, 4)
	//	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACCARD, datasets)
	est := *new(CorrelationEstimator)
	est.datasets = datasets
	est.estType = CorrelationSimilarityTypePearson
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

	newEst := *new(CorrelationEstimator)
	newEst.Deserialize(bytes)
	if est.concurrency != newEst.concurrency {
		t.Log("Concurrency differs")
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

	if newEst.Similarity(datasets[0], datasets[1]) != newEst.SimilarityMatrix().Get(0, 1) {
		t.Log("Something is seriously wrong here", newEst.SimilarityMatrix().Get(0, 1), newEst.Similarity(datasets[0], datasets[1]))
		t.Fail()
	}

}
func TestCorrelations(t *testing.T) {
	getArray := func(d *Dataset) []float64 {
		err := d.ReadFromFile()
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		var array []float64
		for _, tup := range d.Data() {
			array = append(array, tup.Data[0])
		}
		return array
	}
	datasets := createPoolBasedDatasets(10000, 2, 1)
	a, b := getArray(datasets[0]), getArray(datasets[1])
	if math.Abs(Mean(a)-0.5) > 0.2 {
		t.Logf("mean A does not seem correct (%.5f)\n", Mean(a))
		t.Fail()
	}
	if math.Abs(Mean(b)-0.5) > 0.2 {
		t.Logf("mean B does not seem correct (%.5f)\n", Mean(b))
		t.Fail()
	}
	r, s, tau := Pearson(a, b), Spearman(a, b), Kendall(a, b)
	if math.Abs(r) > 1 {
		t.Logf("Pearson r does not seem correct (%.5f)\n", r)
		t.Fail()
	}

	if math.Abs(s) > 1 {
		t.Logf("Spearman s does not seem correct (%.5f)\n", s)
		t.Fail()
	}

	if math.Abs(tau) > 1 {
		t.Logf("Kendall t does not seem correct (%.5f)\n", tau)
		t.Fail()
	}
	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
