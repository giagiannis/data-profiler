package core

import (
	"bytes"
	"errors"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// OrderEstimator estimates the similarity between two different datasets based
// on the ordering of the tuples.
type OrderEstimator struct {
	AbstractDatasetSimilarityEstimator
}

func (e *OrderEstimator) Compute() error {
	e.similarities = NewDatasetSimilarities(len(e.datasets))

	log.Println("Fetching datasets in memory")
	if e.datasets == nil || len(e.datasets) == 0 {
		log.Println("No datasets were given")
		return errors.New("Empty dataset slice")
	}
	for _, d := range e.datasets {
		d.ReadFromFile()
	}

	start := time.Now()
	EstimatorCompute(e)
	e.duration = time.Since(start).Seconds()

	return nil
}

func (e *OrderEstimator) Duration() float64 {
	return e.duration
}

// returns a slice containing the ordering of the tuples
// FIXME: this is HEAVILY under-optimized
func (e *OrderEstimator) getDatasetOrdering(tuples []DatasetTuple) []int {
	// sort dataset
	dt := DatasetTuples(append(make([]DatasetTuple, 0), tuples...))
	sort.Sort(dt)

	// create index slice
	indices := make([]int, len(tuples))
	for i, tup := range tuples {
		idx := 0
		for idx = range dt {
			if tup.Equals(dt[idx]) {
				break
			}
			indices[i] = idx
		}
	}
	return indices
}
func (e *OrderEstimator) Similarity(a, b *Dataset) float64 {
	value, maxDistance := 0.0, 0.0
	o1, o2 := e.getDatasetOrdering(a.Data()), e.getDatasetOrdering(b.Data())
	min := len(o1)

	if len(o1) > len(o2) {
		min = len(o2)
	}

	for i := 0; i < min; i++ {
		value += float64(o1[i]-o2[i]) * float64(o1[i]-o2[i])
		maxDistance += float64(2*i-1-min) * float64(2*i-1-min)
	}
	if value > maxDistance {
		value = maxDistance
	}
	return 1 - math.Sqrt(value/maxDistance)
}

func (e *OrderEstimator) SimilarityMatrix() *DatasetSimilarityMatrix {
	return e.similarities
}

func (e *OrderEstimator) Datasets() []*Dataset {
	return e.datasets
}

func (e *OrderEstimator) Configure(conf map[string]string) {
	if val, ok := conf["concurrency"]; ok {
		conv, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			e.concurrency = int(conv)
		}
	}
}

func (e *OrderEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency": "max num of threads used (int)",
	}
}

func (e *OrderEstimator) SetPopulationPolicy(policy DatasetSimilarityPopulationPolicy) {
	e.popPolicy = policy
}

func (e *OrderEstimator) Serialize() []byte {
	buffer := new(bytes.Buffer)
	buffer.Write(getBytesInt(int(SIMILARITY_TYPE_JACCARD)))
	buffer.Write(getBytesInt(e.concurrency))

	pol := e.popPolicy.Serialize()
	log.Println(len(pol))
	buffer.Write(getBytesInt(len(pol)))
	buffer.Write(pol)

	sim := e.similarities.Serialize()
	log.Println(len(sim))
	buffer.Write(getBytesInt(len(sim)))
	buffer.Write(sim)

	// serialize dataste names
	buffer.Write(getBytesInt(len(e.datasets)))
	for _, d := range e.datasets {
		buffer.WriteString(d.Path() + "\n")
	}

	return buffer.Bytes()
}

func (e *OrderEstimator) Deserialize(b []byte) {
	buffer := bytes.NewBuffer(b)
	tempInt := make([]byte, 4)
	buffer.Read(tempInt) // consume estimator type
	var count int
	buffer.Read(tempInt)
	e.concurrency = getIntBytes(tempInt)

	buffer.Read(tempInt)
	count = getIntBytes(tempInt)
	polBytes := make([]byte, count)
	buffer.Read(polBytes)
	e.popPolicy = *new(DatasetSimilarityPopulationPolicy)
	e.popPolicy.Deserialize(polBytes)

	buffer.Read(tempInt)
	count = getIntBytes(tempInt)
	similarityBytes := make([]byte, count)
	buffer.Read(similarityBytes)
	e.similarities = new(DatasetSimilarityMatrix)
	e.similarities.Deserialize(similarityBytes)

	buffer.Read(tempInt)
	count = getIntBytes(tempInt)
	e.datasets = make([]*Dataset, count)
	for i := range e.datasets {
		line, _ := buffer.ReadString('\n')
		line = strings.TrimSpace(line)
		e.datasets[i] = NewDataset(line)
	}
}
