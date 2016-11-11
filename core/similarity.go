package core

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
)

// DatasetSimilarityEstimator
type DatasetSimilarityEstimator interface {
	Compute() error                        // computes the similarity matrix
	GetSimilarities() *DatasetSimilarities // returns the similarity struct
	Configure(map[string]string)           // provides configuration options
	Options() map[string]string            // list of options for the estimator
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
		a.concurrency = 1
		return a
	} else if estType == BHATTACHARYYA {
		a := new(BhattacharyyaEstimator)
		a.datasets = datasets
		a.concurrency = 1
		a.kdTreeScaleFactor = 0.5
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
	if datasets != nil {
		r.allocateStructs()
	}
	return r
}

func (s *DatasetSimilarities) allocateStructs() {
	s.inverseIndex = make(map[string]int)
	for i := 0; i < len(s.datasets); i++ {
		s.inverseIndex[s.datasets[i].Path()] = i
	}
	s.similarities = make([][]float64, len(s.datasets)-1)
	for i := 0; i < len(s.datasets)-1; i++ {
		s.similarities[i] = make([]float64, len(s.datasets)-i-1)
	}

}

// Set is a setter function for the similarity between two datasets
func (s *DatasetSimilarities) Set(a, b string, value float64) {
	idxA := s.inverseIndex[a]
	idxB := s.inverseIndex[b]
	if idxA == idxB { // do nothing
		return
	} else if idxA > idxB { //we only want to fill the upper diagonal elems
		t := idxB
		idxB = idxA
		idxA = t
	}
	s.similarities[idxA][idxB-idxA-1] = value

}

// Get returns the similarity between two dataset paths
func (s *DatasetSimilarities) Get(a, b string) float64 {
	idxA := s.inverseIndex[a]
	idxB := s.inverseIndex[b]
	if idxA == idxB {
		return 1.0
	} else if idxA > idxB {
		t := idxB
		idxB = idxA
		idxA = t
	}
	return s.similarities[idxA][idxB-idxA-1]
}

func (s DatasetSimilarities) String() string {
	var buf bytes.Buffer
	for i := 0; i < len(s.datasets); i++ {
		for j := 0; j < len(s.datasets); j++ {
			buf.WriteString(fmt.Sprintf("%.5f ",
				s.Get(s.datasets[i].Path(), s.datasets[j].Path())))
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
	buf.Write(getBytesInt(len(s.datasets)))

	for i := 0; i < len(s.datasets); i++ {
		buf.WriteString(fmt.Sprintf("%s\n", s.datasets[i].Path()))
	}

	for i := 0; i < len(s.datasets)-1; i++ {
		for j := 0; j < len(s.datasets)-1-i; j++ {
			buf.Write(getBytesFloat(s.similarities[i][j]))
		}
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
	temp := make([]byte, 4)
	buf.Read(temp)

	s.datasets = make([]*Dataset, getIntBytes(temp))
	for i := range s.datasets {
		tmp, _ := buf.ReadString('\n')
		s.datasets[i] = NewDataset(tmp[:len(tmp)-1])
	}
	s.allocateStructs()

	temp = make([]byte, 8)
	for i := 0; i < len(s.datasets)-1; i++ {
		for j := 0; j < len(s.datasets)-1-i; j++ {
			buf.Read(temp)
			s.similarities[i][j] = getFloatBytes(temp)
		}
	}

	return nil
}

// Datasets method returns the datasets that express the specific similarity
// matrix.
func (s *DatasetSimilarities) Datasets() []*Dataset {
	return s.datasets
}
