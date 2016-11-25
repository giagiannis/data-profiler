package core

import (
	"fmt"
	"log"
	"math/rand"
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
		if sim.datasets[i].Path() != s.datasets[i].Path() {
			log.Printf("Dataset files not identical (%s vs %s)\n",
				sim.datasets[i].Path(), s.datasets[i].Path())
			fmt.Println(len(sim.datasets[i].Path()))
			fmt.Println(len(s.datasets[i].Path()))
			t.FailNow()
		}
	}

	for _, d1 := range datasets {
		for _, d2 := range datasets {
			if sim.Get(d1.Path(), d2.Path()) != s.Get(d1.Path(), d2.Path()) {
				log.Printf("Sim and s provide different values: (%.5f vs %.5f)\n",
					sim.Get(d1.Path(), d2.Path()),
					s.Get(d1.Path(), d2.Path()))
				t.FailNow()
			}
		}
	}
}

func TestDatasetSimilarityClosestIndex(t *testing.T) {
	rand.Seed(int64(time.Now().Nanosecond()))
	datasets := createPoolBasedDatasets(5000, 100, 3)
	sim := NewDatasetSimilarities(datasets)
	chosenIdx := rand.Int() % len(datasets)
	for i, d := range datasets {
		val := rand.Float64()
		if i == chosenIdx {
			val = 1
		}
		sim.Set(datasets[chosenIdx].Path(), d.Path(), val)
	}

	for i := range datasets {
		closest, _ := sim.closestIndex.Get(i)
		if closest != chosenIdx {
			t.Log("Closest idx is not accurate")
			t.FailNow()
		}
	}
}

func TestDatasetSimilaritiesDisabledIndex(t *testing.T) {
	rand.Seed(int64(time.Now().Nanosecond()))
	datasets := createPoolBasedDatasets(5000, 100, 3)
	sim := NewDatasetSimilarities(datasets)
	sim.IndexDisabled(true)
	chosenIdx := rand.Int() % len(datasets)
	for i, d := range datasets {
		val := rand.Float64()
		if i == chosenIdx {
			val = 1
		}
		sim.Set(datasets[chosenIdx].Path(), d.Path(), val)
	}

	for i := range datasets {
		closest, valI := sim.closestIndex.Get(i)
		valA := sim.Get(datasets[chosenIdx].Path(), datasets[i].Path())
		if closest != -1 || valI != -1 || valA <= 0.0 {
			t.Log("Closest idx is wrong")
			t.FailNow()
		}
	}

}
