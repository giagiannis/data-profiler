package core

import (
	"math"
	"testing"
)

func TestKMeansOptions(t *testing.T) {
	a := new(KMeansPartitioner)
	o := a.Options()
	if _, ok := o["k"]; !ok {
		t.Log("k option does not exist")
		t.Fail()
	}
	if _, ok := o["weights"]; !ok {
		t.Log("weights option does not exist")
		t.Fail()
	}
}

func TestKMeansConfigure(t *testing.T) {
	a := new(KMeansPartitioner)
	a.Configure(map[string]string{"k": "10"})
	if a.k != 10 {
		t.Log("k should be 10")
		t.Fail()
	}
	if a.weights != nil {
		t.Log("weights should be nil")
		t.Fail()
	}
	a.Configure(map[string]string{"weights": "0.2,0.5,0.1"})
	if a.k != 1 {
		t.Log("k should be 1")
		t.Fail()
	}
	if a.weights == nil ||
		a.weights[0] != 0.2 ||
		a.weights[1] != 0.5 ||
		a.weights[2] != 0.1 {
		t.Log("weights should be {0.2,0.5,0.1}")
		t.Fail()
	}

}

func TestKMeansConstruct(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 10, 2)
	datasets[0].ReadFromFile()
	a := new(KMeansPartitioner)
	a.Configure(map[string]string{"k": "5"})
	a.Construct(datasets[0].Data())
	if a.centroids == nil {
		t.Log("Centroids are null")
		t.Fail()
	}
}
func TestKMeansPartition(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 10, 2)
	datasets[0].ReadFromFile()
	a := new(KMeansPartitioner)

	datasets[1].ReadFromFile()
	clusters, err := a.Partition(datasets[1].Data())
	if err == nil || clusters != nil {
		t.Log("should return an error")
		t.Fail()
	}

	a.Configure(map[string]string{"k": "5"})
	a.Construct(datasets[0].Data())

	clusters, err = a.Partition(nil)
	if err == nil || clusters != nil {
		t.Log("should return an error")
		t.Fail()
	}

	other := createPoolBasedDatasets(10000, 1, 3)[0]
	other.ReadFromFile()
	clusters, err = a.Partition(other.Data())
	if err == nil || clusters != nil {
		t.Log("should return an error")
		t.Fail()
	}

	clusters, err = a.Partition(datasets[1].Data())
	if err != nil || clusters == nil {
		t.Log("should not return an error")
		t.Fail()
	}
	if len(clusters) != 5 {
		t.Log("wrong number of clusters")
		t.Fail()
	}
}

func TestKMeansEstimateWeights(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 10, 2)
	datasets[0].ReadFromFile()
	a := new(KMeansPartitioner)
	a.Configure(map[string]string{"k": "5"})
	a.estimateWeights(datasets[0].Data())
	if len(a.weights) != len(datasets[0].Data()[0].Data) {
		t.Log("no weights returned")
		t.Fail()
	}
	for _, v := range a.weights {
		if math.IsInf(v, 0) || v < 0.0 {
			t.Log("weight equals infty or negative")
			t.Fail()
		}
	}
}

func TestKMeansInitializeCentroids(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 10, 2)
	datasets[0].ReadFromFile()

	a := new(KMeansPartitioner)
	a.Configure(map[string]string{"k": "5"})
	a.estimateWeights(datasets[0].Data())
	a.initializeCentroids(datasets[0].Data())
	if a.centroids == nil {
		t.Log("centroids found nil")
		t.Fail()
	}
	for _, c := range a.centroids {
		if c.Data == nil {
			t.Log("nil centroid found")
			t.Fail()
		}
		if len(c.Data) != len(datasets[0].Data()[0].Data) {
			t.Log("centroid has wrong # of dimensions")
			t.Fail()
		}
	}
}

func TestKMeansAssignTuplesToCentroids(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 10, 2)
	datasets[0].ReadFromFile()

	a := new(KMeansPartitioner)
	a.Configure(map[string]string{"k": "5"})
	a.estimateWeights(datasets[0].Data())
	a.initializeCentroids(datasets[0].Data())
	clusters := a.assignTuplesToCentroids(datasets[0].Data())
	if len(clusters) != len(a.centroids) {
		t.Log("wrong number of clusters")
		t.Fail()
	}
}

func TestKMeansEstimateCentroids(t *testing.T) {
	datasets := createPoolBasedDatasets(100, 1, 2)
	datasets[0].ReadFromFile()

	a := new(KMeansPartitioner)
	a.Configure(map[string]string{"k": "10"})
	a.estimateWeights(datasets[0].Data())
	a.initializeCentroids(datasets[0].Data())
	clusters := a.assignTuplesToCentroids(datasets[0].Data())
	newCentroids := a.estimateCentroids(clusters)
	if len(newCentroids) != len(a.centroids) {
		t.Log("wrong number of clusters")
		t.Fail()
	}
}

func TestKMeansDeSerialize(t *testing.T) {
	datasets := createPoolBasedDatasets(100, 1, 2)
	datasets[0].ReadFromFile()

	a := new(KMeansPartitioner)
	a.Configure(map[string]string{"k": "10"})
	a.Construct(datasets[0].Data())

	buff := a.Serialize()

	b := new(KMeansPartitioner)
	b.Deserialize(buff)
	if a.k != b.k {
		t.Log("k values do not match")
		t.FailNow()
	}
	for i := range a.weights {
		if b.weights==nil || len(b.weights) < i || a.weights[i] != b.weights[i] {
			t.Log("weights do not match")
			t.FailNow()
		}
	}
	for i := range a.centroids {
		if b.centroids == nil || len(b.centroids) < i {
			t.Log("centroids do not match")
			t.FailNow()
		} else {
			for j := range a.centroids[i].Data {
				if b.centroids == nil || len(b.centroids[i].Data) < j ||
					a.centroids[i].Data[j] != b.centroids[i].Data[j] {
					t.Log("centroids content do not match")
					t.FailNow()
				}
			}
		}
	}

}
