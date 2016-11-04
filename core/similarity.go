package core

// DatasetSimilarityEstimator
type DatasetSimilarityEstimator interface {
	Compute() error                        // computes the similarity matrix
	GetSimilarities() *DatasetSimilarities // returns the similarity struct
}

type DatasetSimilarityEstimatorType uint

const (
	JACOBBI DatasetSimilarityEstimatorType = iota + 1
	BHATTACHARYYA
)

// Factory method for creating a DatasetSimilarityEstimator
func NewDatasetSimilarityEstimator(
	estType DatasetSimilarityEstimatorType,
	datasets []*Dataset) DatasetSimilarityEstimator {
	if estType == JACOBBI {
		a := new(JacobbiEstimator)
		a.datasets = datasets
		a.concurrency = 8
		return a
	} else if estType == BHATTACHARYYA {
		a := new(BhattacharyyaEstimator)
		a.datasets = datasets
		a.concurrency = 8
		return a
	}

	return nil
}

// DatasetSimilarities represent the struct that holds the results of  a
// dataset similarity estimation. It also provides the necessary
type DatasetSimilarities struct {
	datasets     []*Dataset     // the datasets slice
	inverseIndex map[string]int // the inverse index
	similarities [][]float64    // the actual similarities holder
}

// NewDatasetSimilarities is the constructor for the DatasetSimilarities struct
func NewDatasetSimilarities(datasets []*Dataset) *DatasetSimilarities {
	r := new(DatasetSimilarities)
	r.datasets = datasets
	r.inverseIndex = make(map[string]int)
	for i := 0; i < len(r.datasets); i++ {
		r.inverseIndex[r.datasets[i].Id()] = i
	}
	r.similarities = make([][]float64, len(r.datasets)-1)
	for i := 0; i < len(r.datasets)-1; i++ {
		r.similarities[i] = make([]float64, len(r.datasets)-i-1)
	}
	return r
}

// SetSimilarity is a setter function for the similarity between two datasets
func (s *DatasetSimilarities) Set(a, b Dataset, value float64) {
	idxA := s.inverseIndex[a.Id()]
	idxB := s.inverseIndex[b.Id()]
	if idxA == idxB { // do nothing
	} else if idxA > idxB { //we only want to fill the upper diagonal elems
		t := idxB
		idxB = idxA
		idxA = t
	}
	s.similarities[idxA][idxB-idxA-1] = value

}

// GetSimilarity is a getter function for the simil
func (s *DatasetSimilarities) Get(a, b Dataset) float64 {
	idxA := s.inverseIndex[a.Id()]
	idxB := s.inverseIndex[b.Id()]
	if idxA == idxB {
		return 1.0
	} else if idxA > idxB {
		t := idxB
		idxB = idxA
		idxA = t
	}
	return s.similarities[idxA][idxB-idxA-1]
}
