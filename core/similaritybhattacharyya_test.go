package core

import (
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestBhattacharyyaCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(20000, 20, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_BHATTACHARYYA, datasets)
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
	for i := range datasets {
		for j := range datasets {
			v1 := s.Get(i, j)
			v2 := s.Get(j, i)
			if v1 != v2 {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
			if v1 < 0 || v1 > 1.0 {
				t.Log("Similarity value out of range [0,1]", v1, v2)
				t.FailNow()
			}
			if v1 == 0 {
				t.Log("Zero element found (?)")
			}
			if v1 == 1 && i != j {
				t.Log("Similarity of 1 in non-diagonal element")
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

func TestBhattacharyyaComputeAppxCnt(t *testing.T) {
	rand.Seed(int64(time.Now().Nanosecond()))
	datasets := createPoolBasedDatasets(200, 200, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_BHATTACHARYYA, datasets)
	conf := map[string]string{"concurrency": "10"}
	est.Configure(conf)
	policy := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"count": 20,
		},
	}
	est.PopulationPolicy(policy)
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
	for i := range datasets {
		for j := range datasets {
			v1 := s.Get(i, j)
			v2 := s.Get(j, i)
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

func TestBhattacharyyaComputeAppxThres(t *testing.T) {
	rand.Seed(int64(time.Now().Nanosecond()))
	datasets := createPoolBasedDatasets(200, 200, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_BHATTACHARYYA, datasets)
	conf := map[string]string{"concurrency": "10"}
	est.Configure(conf)
	policy := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"threshold": 0.985,
		},
	}
	est.PopulationPolicy(policy)
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
	for i := range datasets {
		for j := range datasets {
			v1 := s.Get(i, j)
			v2 := s.Get(j, i)
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

func TestKdtreeNodeSerialization(t *testing.T) {
	datasets := createPoolBasedDatasets(50000, 1, 4)
	datasets[0].ReadFromFile()
	tree := NewKDTreePartition(datasets[0].Data())
	tree.Prune(tree.MinHeight() / 2)
	b := tree.Serialize()
	newTree := new(kdTreeNode)
	newTree.Deserialize(b)
	var dfsTraversal func(treeA, treeB *kdTreeNode) bool
	dfsTraversal = func(treeA, treeB *kdTreeNode) bool {
		if treeA == nil && treeB == nil {
			return true
		}
		if treeA.dim != treeB.dim || treeA.value != treeB.value {
			return false
		} else {
			return dfsTraversal(treeA.left, treeB.left) && dfsTraversal(treeA.right, treeB.right)
		}
	}
	if !dfsTraversal(tree, newTree) {
		t.Log("Trees not equal")
		t.FailNow()
	}
}
