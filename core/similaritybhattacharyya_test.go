package core

import (
	"log"
	"os"
	"testing"
)

func TestBhattacharyyaCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(20000, 20, 4)
	est := NewDatasetSimilarityEstimator(BHATTACHARYYA, datasets)
	conf := map[string]string{"concurrency": "10"}
	est.Configure(conf)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	s := est.GetSimilarities()
	if s == nil {
		t.Log("Nil similarities returned")
		t.FailNow()
	}
	for _, d1 := range datasets {
		for _, d2 := range datasets {
			v1 := s.Get(d1.Path(), d2.Path())
			v2 := s.Get(d2.Path(), d1.Path())
			if v1 != v2 {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
			if v1 < 0 || v1 > 1.0 {
				t.Log("Similarity value out of range [0,1]", v1, v2)
				t.FailNow()
			}
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestKdTree(t *testing.T) {
	dataset := createPoolBasedDatasets(20000, 1, 5)[0]
	err := dataset.ReadFromFile()
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	kd := NewKDTreePartition(dataset.Data())
	kd.Prune(kd.Height() / 5)
	ids := kd.GetLeafIndex(dataset.Data())
	sum := 0
	for _, v := range ids {
		sum += v
	}
	if sum != len(dataset.Data()) {
		log.Println("Not all tuples are indexed!!")
		t.FailNow()
	}
	os.Remove(dataset.Path())

}
