package core

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
)

// DatasetSimilarityEstimator
type DatasetSimilarityEstimator interface {
	Compute() error // computes the similarity matrix
	//	Similarity(a, b *Dataset)                           // returns the similarity for 2 datasets
	GetSimilarities() *DatasetSimilarities              // returns the similarity struct
	Configure(map[string]string)                        // provides configuration options
	Options() map[string]string                         // list of options for the estimator
	PopulationPolicy(DatasetSimilarityPopulationPolicy) // sets the population policy for the estimator
}

type DatasetSimilarityEstimatorType uint

const (
	SIMILARITY_TYPE_JACOBBI       DatasetSimilarityEstimatorType = iota
	SIMILARITY_TYPE_BHATTACHARYYA DatasetSimilarityEstimatorType = iota + 1
	SIMILARITY_TYPE_SCRIPT        DatasetSimilarityEstimatorType = iota + 2
)

func (t DatasetSimilarityEstimatorType) String() string {
	if t == SIMILARITY_TYPE_JACOBBI {
		return "Jacobbi"
	} else if t == SIMILARITY_TYPE_BHATTACHARYYA {
		return "Bhattacharyya"
	} else if t == SIMILARITY_TYPE_SCRIPT {
		return "Script"
	}
	return ""
}

type DatasetSimilarityPopulationPolicy struct {
	PolicyType DatasetSimilarityPopulationPolicyType
	Parameters map[string]float64
}

type DatasetSimilarityPopulationPolicyType uint

const (
	// FULL policy needs no params
	POPULATION_POL_FULL DatasetSimilarityPopulationPolicyType = iota
	// APRX must have defined one of two params: count (how many points)
	// or threshold (percentage in similarity gain)
	POPULATION_POL_APRX DatasetSimilarityPopulationPolicyType = iota + 1
)

// Factory method for creating a DatasetSimilarityEstimator
func NewDatasetSimilarityEstimator(
	estType DatasetSimilarityEstimatorType,
	datasets []*Dataset) DatasetSimilarityEstimator {
	policy := *new(DatasetSimilarityPopulationPolicy)
	policy.PolicyType = POPULATION_POL_FULL
	if estType == SIMILARITY_TYPE_JACOBBI {
		a := new(JacobbiEstimator)
		a.PopulationPolicy(policy)
		a.datasets = datasets
		a.concurrency = 1
		return a
	} else if estType == SIMILARITY_TYPE_BHATTACHARYYA {
		a := new(BhattacharyyaEstimator)
		a.PopulationPolicy(policy)
		a.datasets = datasets
		a.concurrency = 1
		a.kdTreeScaleFactor = 0.5
		return a
	} else if estType == SIMILARITY_TYPE_SCRIPT {
		a := new(ScriptSimilarityEstimator)
		a.PopulationPolicy(policy)
		a.datasets = datasets
		a.concurrency = 1
		a.normDegree = 1
		return a
	}
	return nil
}

// DatasetSimilarities represent the struct that holds the results of  a
// dataset similarity estimation. It also provides the necessary
type DatasetSimilarities struct {
	similarities  [][]float64   // the actual similarities holder
	indexDisabled bool          // indicates whether the closestIndex is disabled or not
	closestIndex  *closestIndex // index that hold the closest datasets
	capacity      int           // represents the capacity of the sim matrix
}

// NewDatasetSimilarities is the constructor for the DatasetSimilarities struct,
// expecting the number of datasets that will be held by it. If capacity=0, this
// implies that the Similarity Matrix will be deserialzed.
func NewDatasetSimilarities(capacity int) *DatasetSimilarities {
	r := new(DatasetSimilarities)
	r.indexDisabled = false
	r.capacity = capacity
	if capacity != 0 {
		r.allocateStructs()
	}
	return r
}

// IndexDisabled sets whether the closest dataset index should be disabled or not.
// The index is useless if the FULL Estimator strategy is being followed.
func (s *DatasetSimilarities) IndexDisabled(flag bool) {
	s.indexDisabled = flag
}

// NumberOfFullNodes returns the number of nodes the similarity of which
// has been calculated for all the nodes. This number can work as a measure
// of how close to the full similarity matrix the current object is.
func (s *DatasetSimilarities) FullyCalculatedNodes() int {
	if s.indexDisabled {
		return s.capacity
	}
	count := 0
	for i := 0; i < s.capacity; i++ {
		if idx, _ := s.closestIndex.Get(i); idx == i {
			count += 1
		}
	}
	return count
}

func (s *DatasetSimilarities) allocateStructs() {
	s.similarities = make([][]float64, s.capacity-1)
	for i := 0; i < s.capacity-1; i++ {
		s.similarities[i] = make([]float64, s.capacity-i-1)
	}
	s.closestIndex = newClosestIndex(s.capacity)
}

func (s *DatasetSimilarities) Capacity() int {
	return s.capacity
}

// Set is a setter function for the similarity between two datasets
func (s *DatasetSimilarities) Set(idxA, idxB int, value float64) {
	if idxA == idxB { // do nothing
		if !s.indexDisabled {
			s.closestIndex.CheckAndSet(idxA, idxB, value)
			s.closestIndex.CheckAndSet(idxB, idxA, value)
		}
		return
	} else if idxA > idxB { //we only want to fill the upper diagonal elems
		t := idxB
		idxB = idxA
		idxA = t
	}
	s.similarities[idxA][idxB-idxA-1] = value
	if !s.indexDisabled {
		s.closestIndex.CheckAndSet(idxA, idxB, value)
		s.closestIndex.CheckAndSet(idxB, idxA, value)
	}
}

// Get returns the similarity between two dataset paths
func (s *DatasetSimilarities) Get(idxA, idxB int) float64 {
	if !s.indexDisabled {
		idxA, _ = s.closestIndex.Get(idxA)
		idxB, _ = s.closestIndex.Get(idxB)
	}
	if idxA == idxB {
		return 1.0
	} else if idxA > idxB {
		t := idxB
		idxB = idxA
		idxA = t
	}
	return s.similarities[idxA][idxB-idxA-1]
}

// LeastSimilar method returns the dataset that presents the lowest
// similarity among the examined datasets
func (s *DatasetSimilarities) LeastSimilar() (int, float64) {
	return s.closestIndex.LeastSimilar()
}

func (s DatasetSimilarities) String() string {
	var buf bytes.Buffer
	for i := 0; i < s.capacity; i++ {
		for j := 0; j < s.capacity; j++ {
			buf.WriteString(fmt.Sprintf("%.5f ", s.Get(i, j)))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// Serialize method returns a byte slice that represents the similarity matrix
func (s *DatasetSimilarities) Serialize() []byte {

	getBytesInt := func(val int) []byte {
		temp := make([]byte, 4)
		binary.BigEndian.PutUint32(temp, uint32(val))
		return temp
	}

	getBytesFloat := func(val float64) []byte {
		bits := math.Float64bits(val)
		temp := make([]byte, 8)
		binary.BigEndian.PutUint64(temp, bits)
		return temp
	}

	buf := new(bytes.Buffer)
	buf.Write(getBytesInt(s.capacity))
	for i := 0; i < s.capacity-1; i++ {
		for j := 0; j < s.capacity-1-i; j++ {
			buf.Write(getBytesFloat(s.similarities[i][j]))
		}
	}

	for i := 0; i < s.capacity; i++ {
		buf.Write(getBytesFloat(float64(s.closestIndex.closestIdx[i])))
		buf.Write(getBytesFloat(s.closestIndex.similarity[i]))
	}
	if s.indexDisabled {
		buf.WriteString("1")
	} else {
		buf.WriteString("0")
	}

	// compress before you send
	var compressed bytes.Buffer
	wr, err := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
	//	wr := gzip.NewWrite(&compressed)
	defer wr.Close()
	if err != nil {
		log.Println("Error message from compression: ", err)
	}
	wr.Write(buf.Bytes())
	wr.Flush()
	log.Println("Compressed bytes", len(compressed.Bytes()), len(buf.Bytes()))
	return compressed.Bytes()
	//	return buf.Bytes()
}

// Deserialize instantiates an empty DatasetSimilarities object. In case of
// parse failure, an error is thrown
func (s *DatasetSimilarities) Deserialize(buff []byte) error {
	// decompress stream
	re, err := gzip.NewReader(bytes.NewBuffer(buff))
	if err != nil {
		log.Println("Error message from compression: ", err)
	}
	defer re.Close()
	buff, _ = ioutil.ReadAll(re)

	getIntBytes := func(buf []byte) int {
		return int(binary.BigEndian.Uint32(buf))
	}
	getFloatBytes := func(buf []byte) float64 {
		bits := binary.BigEndian.Uint64(buf)
		float := math.Float64frombits(bits)
		return float
	}

	buf := bytes.NewBuffer(buff)
	tempInt := make([]byte, 4)
	buf.Read(tempInt)
	s.capacity = getIntBytes(tempInt)
	s.allocateStructs()

	tempFloat := make([]byte, 8)
	for i := 0; i < s.capacity-1; i++ {
		for j := 0; j < s.capacity-1-i; j++ {
			buf.Read(tempFloat)
			s.similarities[i][j] = getFloatBytes(tempFloat)
		}
	}

	for i := 0; i < s.capacity; i++ {
		buf.Read(tempFloat)
		s.closestIndex.closestIdx[i] = int(getFloatBytes(tempFloat))
		buf.Read(tempFloat)
		s.closestIndex.similarity[i] = getFloatBytes(tempFloat)
	}
	if val, _ := buf.ReadString('\n'); val == "1" {
		s.indexDisabled = true
	} else {
		s.indexDisabled = false
	}
	return nil
}

// datasets along with their respective similarities
type closestIndex struct {
	closestIdx []int
	similarity []float64
}

func newClosestIndex(datasets int) *closestIndex {
	res := new(closestIndex)
	res.closestIdx = make([]int, datasets)
	res.similarity = make([]float64, datasets)
	for i := range res.closestIdx {
		res.closestIdx[i] = -1   // represents a NUL dataset index
		res.similarity[i] = -1.0 // represents a NUL similarity
	}
	return res
}

// Returns the index and similarity of the most similar dataset
func (s *closestIndex) Get(idx int) (int, float64) {
	if idx < len(s.closestIdx) {
		return s.closestIdx[idx], s.similarity[idx]
	}
	return -1, -1.0
}

// Sets the index and the similarity of the most similar dataset
func (s *closestIndex) Set(srcIdx, dstIdx int, similarity float64) {
	s.closestIdx[srcIdx] = dstIdx
	s.similarity[srcIdx] = similarity
}

// Sets the index and the similarity of the most similar dataset, iff the provided similarity
// is higher than the one previously stored
func (s *closestIndex) CheckAndSet(srcIdx, dstIdx int, similarity float64) {
	if s.similarity[srcIdx] < similarity {
		s.closestIdx[srcIdx] = dstIdx
		s.similarity[srcIdx] = similarity
	}
}

// Returns the dataset index (and its respective value) with the lowest similarity
// to its most close dataset
func (s *closestIndex) LeastSimilar() (int, float64) {
	minIdx, minV := 0, s.similarity[0]
	for i, v := range s.similarity {
		if v < minV {
			minV = v
			minIdx = i
		} else if v == minV && rand.Int()%2 == 0 { // random index
			minIdx = i
		}
	}
	return minIdx, minV
}
