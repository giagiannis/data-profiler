package core

import (
	"log"
	"math/rand"
	"os"
	"testing"
)

func TestDatasetSimilarityPolicySerialization(t *testing.T) {
	pol := new(DatasetSimilarityPopulationPolicy)
	pol.Parameters = map[string]float64{"count": 10.0, "foo": 100.0}
	pol.PolicyType = POPULATION_POL_APRX
	bytes := pol.Serialize()
	newPol := new(DatasetSimilarityPopulationPolicy)
	newPol.Deserialize(bytes)
	if newPol.PolicyType != pol.PolicyType {
		t.Log("Policy types different")
		t.Fail()
	}
	for k, v := range pol.Parameters {
		val, ok := newPol.Parameters[k]
		if !ok || val != v {
			t.Log("Missing key or different value", ok, val, v)
			t.Fail()
		}
	}
}

func TestDatasetSerialize(t *testing.T) {
	datasets := createPoolBasedDatasets(5000, 10, 3)
	e := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_BHATTACHARYYA, datasets)
	e.Compute()
	s := e.SimilarityMatrix()
	b := s.Serialize()

	sim := new(DatasetSimilarityMatrix)
	sim.Deserialize(b)

	for i := range datasets {
		if sim.closestIndex.similarity[i] != s.closestIndex.similarity[i] ||
			sim.closestIndex.closestIdx[i] != s.closestIndex.closestIdx[i] {
			log.Println("closestIndex contains different elements at dataset", i)
			log.Println(sim.closestIndex, s.closestIndex)
			t.FailNow()
		}
	}

	for i := range datasets {
		for j := range datasets {
			if sim.Get(i, j) != s.Get(j, i) {
				log.Printf("Sim and s provide different values: (%.5f vs %.5f)\n",
					sim.Get(i, j),
					s.Get(i, j))
				t.FailNow()
			}
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestDatasetSimilarityClosestIndex(t *testing.T) {
	datasets := createPoolBasedDatasets(5000, 100, 3)
	sim := NewDatasetSimilarities(len(datasets))
	chosenIdx := rand.Int() % len(datasets)
	for i := range datasets {
		val := rand.Float64()
		if i == chosenIdx {
			val = 1
		}
		sim.Set(chosenIdx, i, val)
	}

	for i := range datasets {
		closest, _ := sim.closestIndex.Get(i)
		if closest != chosenIdx {
			t.Log("Closest idx is not accurate")
			t.FailNow()
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestDatasetSimilaritiesDisabledIndex(t *testing.T) {
	datasets := createPoolBasedDatasets(5000, 100, 3)
	sim := NewDatasetSimilarities(len(datasets))
	sim.IndexDisabled(true)
	chosenIdx := rand.Int() % len(datasets)
	for i := range datasets {
		val := rand.Float64()
		if i == chosenIdx {
			val = 1
		}
		sim.Set(chosenIdx, i, val)
	}

	for i := range datasets {
		closest, valI := sim.closestIndex.Get(i)
		valA := sim.Get(chosenIdx, i)
		if closest != -1 || valI != -1 || valA <= 0.0 {
			t.Log("Closest idx is wrong")
			t.FailNow()
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestDeserializeSimilarityEstimator(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 10, 2)
	// jaccard
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACCARD, datasets)
	est.Configure(map[string]string{"concurrency": "10"})
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		buf := est.Serialize()
		newEst := DeserializeSimilarityEstimator(buf)
		idxA, idxB := rand.Intn(len(datasets)), rand.Intn(len(datasets))
		a, b := datasets[idxA], datasets[idxB]
		if newEst.Similarity(a, b) != est.Similarity(a, b) ||
			newEst.Similarity(a, b) != est.SimilarityMatrix().Get(idxA, idxB) {
			t.Log("Deserialization error")
			t.Fail()
		}
	}
	// bhat
	est = NewDatasetSimilarityEstimator(SIMILARITY_TYPE_BHATTACHARYYA, datasets)
	est.Configure(map[string]string{"concurrency": "10"})
	err = est.Compute()
	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		buf := est.Serialize()
		newEst := DeserializeSimilarityEstimator(buf)
		idxA, idxB := rand.Intn(len(datasets)), rand.Intn(len(datasets))
		a, b := datasets[idxA], datasets[idxB]
		if newEst.Similarity(a, b) != est.Similarity(a, b) ||
			newEst.Similarity(a, b) != est.SimilarityMatrix().Get(idxA, idxB) {
			t.Log("Deserialization error")
			t.Fail()
		}
	}

	// script
	est = NewDatasetSimilarityEstimator(SIMILARITY_TYPE_SCRIPT, datasets)
	est.Configure(map[string]string{"concurrency": "10", "script": analysisScript})
	err = est.Compute()
	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		buf := est.Serialize()
		newEst := DeserializeSimilarityEstimator(buf)
		idxA, idxB := rand.Intn(len(datasets)), rand.Intn(len(datasets))
		a, b := datasets[idxA], datasets[idxB]
		if newEst.Similarity(a, b) != est.Similarity(a, b) ||
			newEst.Similarity(a, b) != est.SimilarityMatrix().Get(idxA, idxB) {
			t.Log("Deserialization error")
			t.Fail()
		}
	}
}
