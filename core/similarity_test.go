package core

import (
	"fmt"
	"log"
	"testing"
)

func TestDatasetSerialize(t *testing.T) {
	datasets := createPoolBasedDatasets(5000, 100, 3)
	e := NewDatasetSimilarityEstimator(BHATTACHARYYA, datasets)
	e.Compute()
	s := e.GetSimilarities()
	b := s.Serialize()

	sim := new(DatasetSimilarities)
	sim.Deserialize(b)

	for i := range datasets {
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
				log.Printf("(%.5f vs %.5f)\n",
					sim.Get(d1.Path(), d2.Path()),
					s.Get(d1.Path(), d2.Path()))
				//				t.FailNow()
			}
		}
	}
}
