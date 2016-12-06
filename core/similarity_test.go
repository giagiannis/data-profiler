package core

import (
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestDatasetSerialize(t *testing.T) {
	datasets := createPoolBasedDatasets(5000, 10, 3)
	e := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_BHATTACHARYYA, datasets)
	e.Compute()
	s := e.GetSimilarities()
	b := s.Serialize()

	sim := new(DatasetSimilarities)
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
	rand.Seed(int64(time.Now().Nanosecond()))
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
	rand.Seed(int64(time.Now().Nanosecond()))
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
