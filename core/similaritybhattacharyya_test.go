package core

import (
	"log"
	"os"
	"testing"
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
	datasets := createPoolBasedDatasets(10000, 1, 2)
	datasets[0].ReadFromFile()
	tree := NewKDTreePartition(datasets[0].Data())
	tree.Prune(tree.MinHeight() * 4 / 5)
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

func TestBhattacharyyaSerialization(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 100, 4)
	//	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACOBBI, datasets)
	est := *new(BhattacharyyaEstimator)
	est.datasets = datasets
	est.kdTreeScaleFactor = 0.75
	est.concurrency = 10
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_FULL,
		Parameters: map[string]float64{},
	}
	est.PopulationPolicy(pol)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	bytes := est.Serialize()

	newEst := *new(BhattacharyyaEstimator)
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

	for k, v := range est.inverseIndex {
		if newEst.inverseIndex[k] != v {
			t.Log("Inverse index failed")
			t.Fail()
		}
	}

	for i, arr := range est.pointsPerRegion {
		for j, v := range arr {
			if v != newEst.pointsPerRegion[i][j] {
				t.Log("Points per region not the same", i, j)
				t.Fail()
			}
		}
	}

	for i, v := range est.datasetsSize {
		if v != newEst.datasetsSize[i] {
			t.Log("Datasets sizes not the same", i)
			t.Fail()
		}
	}

	if newEst.Similarity(datasets[0], datasets[1]) != newEst.GetSimilarities().Get(0, 1) {
		t.Log("Something is seriously wrong here", newEst.GetSimilarities().Get(0, 1), newEst.Similarity(datasets[0], datasets[1]))
		t.Fail()
	}
}
