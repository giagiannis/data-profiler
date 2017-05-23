package core

import (
	"log"
	"math"
	"testing"
)

func TestBhattacharyyaCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(20000, 20, 4)
	est := NewDatasetSimilarityEstimator(SimilarityTypeBhattacharyya, datasets)
	conf := map[string]string{"concurrency": "10"}
	est.Configure(conf)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.SimilarityMatrix()
	smSanityCheck(s, t)
	cleanDatasets(datasets)
}

func TestKdTree(t *testing.T) {
	datasets := createPoolBasedDatasets(20000, 1, 5)
	dataset := datasets[0]
	err := dataset.ReadFromFile()
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	kd := newKDTreePartition(dataset.Data())
	kd.Prune(kd.Height() - 2)
	ids := kd.GetLeafIndex(dataset.Data())
	sum := 0
	for _, v := range ids {
		sum += v
	}
	if sum != len(dataset.Data()) {
		log.Println("Not all tuples are indexed!!")
		t.FailNow()
	}
	cleanDatasets(datasets)

}

func TestKdPruning(t *testing.T) {
	datasets := createPoolBasedDatasets(20000, 1, 5)
	dataset := datasets[0]
	err := dataset.ReadFromFile()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	height := int(math.Log2(float64(len(dataset.Data())))) + 1
	for i := 0; i < height; i++ {
		kd := newKDTreePartition(dataset.Data())
		if i < kd.Height() {
			kd.Prune(kd.Height() - i)
			ids := kd.GetLeafIndex(dataset.Data())
			sum := .0
			for _, v := range ids {
				sum += math.Sqrt(float64(v * v))
			}
			if sum != math.Sqrt(float64(len(dataset.Data())*len(dataset.Data()))) {
				t.FailNow()
			}
		}
	}

	cleanDatasets(datasets)
}

func TestBhattacharyyaComputeAppxCnt(t *testing.T) {
	datasets := createPoolBasedDatasets(200, 200, 4)
	est := NewDatasetSimilarityEstimator(SimilarityTypeBhattacharyya, datasets)
	conf := map[string]string{"concurrency": "10"}
	est.Configure(conf)
	policy := DatasetSimilarityPopulationPolicy{
		PolicyType: PopulationPolicyAprx,
		Parameters: map[string]float64{
			"count": 20,
		},
	}
	est.SetPopulationPolicy(policy)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	s := est.SimilarityMatrix()
	smSanityCheck(s, t)
	cleanDatasets(datasets)
}

func TestBhattacharyyaComputeAppxThres(t *testing.T) {
	datasets := createPoolBasedDatasets(200, 200, 4)
	est := NewDatasetSimilarityEstimator(SimilarityTypeBhattacharyya, datasets)
	conf := map[string]string{"concurrency": "10"}
	est.Configure(conf)
	policy := DatasetSimilarityPopulationPolicy{
		PolicyType: PopulationPolicyAprx,
		Parameters: map[string]float64{
			"threshold": 0.995,
		},
	}
	est.SetPopulationPolicy(policy)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	s := est.SimilarityMatrix()
	smSanityCheck(s, t)
	cleanDatasets(datasets)
}

func TestKdtreeNodeSerialization(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 1, 2)
	datasets[0].ReadFromFile()
	tree := newKDTreePartition(datasets[0].Data())
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
		}
		return dfsTraversal(treeA.left, treeB.left) && dfsTraversal(treeA.right, treeB.right)
	}
	if !dfsTraversal(tree, newTree) {
		t.Log("Trees not equal")
		t.FailNow()
	}
	cleanDatasets(datasets)
}

func TestBhattacharyyaSerialization(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 4)
	//	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACCARD, datasets)
	est := *new(BhattacharyyaEstimator)
	est.datasets = datasets
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: PopulationPolicyFull,
		Parameters: map[string]float64{},
	}
	est.SetPopulationPolicy(pol)
	est.Configure(map[string]string{"concurrency": "10", "tree.scale": "0.75"})
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	bytes := est.Serialize()

	newEst := *new(BhattacharyyaEstimator)
	newEst.Deserialize(bytes)

	estimatorsCheck(newEst.AbstractDatasetSimilarityEstimator,
		est.AbstractDatasetSimilarityEstimator, t)
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

	cleanDatasets(datasets)
}
