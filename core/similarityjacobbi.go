package core

import (
	"bytes"
	"errors"
	"log"
	"strconv"
	"strings"
)

// JacobbiEstimator estimates the Jacobbi coefficients between the different
// datasets. The Jacobbi coefficient between two datasets is defined as
// the cardinality of the intersection divided by the cardinality of the
// union of the two datasets.
type JacobbiEstimator struct {
	datasets    []*Dataset                        // the slice of datasets
	concurrency int                               // max threads running in parallel
	popPolicy   DatasetSimilarityPopulationPolicy // the policy with which the similarities matrix will be populated

	similarities *DatasetSimilarities // holds the similarities
}

func (e *JacobbiEstimator) Compute() error {
	e.similarities = NewDatasetSimilarities(len(e.datasets))

	log.Println("Fetching datasets in memory")
	if e.datasets == nil || len(e.datasets) == 0 {
		log.Println("No datasets were given")
		return errors.New("Empty dataset slice")
	}
	for _, d := range e.datasets {
		d.ReadFromFile()
	}
	if e.popPolicy.PolicyType == POPULATION_POL_FULL {
		e.similarities.IndexDisabled(true) // I don't need the index
		log.Printf("Starting Jacobbi computation (%d threads)", e.concurrency)
		c := make(chan bool, e.concurrency)
		done := make(chan bool)
		for j := 0; j < e.concurrency; j++ {
			c <- true
		}
		for i := 0; i < len(e.datasets)-1; i++ {
			go func(c, done chan bool, i int) {
				<-c
				e.calculateLine(i, i)
				c <- true
				done <- true
			}(c, done, i)
		}
		for j := 0; j < len(e.datasets)-1; j++ {
			<-done
		}
		log.Println("Done")
	} else if e.popPolicy.PolicyType == POPULATION_POL_APRX {
		e.similarities.IndexDisabled(false) // I need the index
		if count, ok := e.popPolicy.Parameters["count"]; ok {
			log.Printf("Fixed number of points execution (count: %.0f)\n", count)
			for i := 0.0; i < count; i++ {
				idx, val := e.similarities.LeastSimilar()
				log.Println("Computing the similarities for ", idx, val)
				e.calculateLine(0, idx)
			}

		} else if threshold, ok := e.popPolicy.Parameters["threshold"]; ok {
			log.Printf("Threshold based execution (threshold: %.5f)\n", threshold)
			idx, val := e.similarities.LeastSimilar()
			for val < threshold {
				log.Printf("Computing the similarities for (%d, %.5f)\n", idx, val)
				e.calculateLine(0, idx)
				idx, val = e.similarities.LeastSimilar()
			}

		}
	}

	return nil
}
func (e *JacobbiEstimator) Similarity(a, b *Dataset) float64 {
	inter := len(DatasetsIntersection(a, b))
	union := len(DatasetsUnion(a, b))
	value := float64(inter) / float64(union)
	return value
}

func (e *JacobbiEstimator) GetSimilarities() *DatasetSimilarities {
	return e.similarities
}

func (e *JacobbiEstimator) Datasets() []*Dataset {
	return e.datasets
}

func (e *JacobbiEstimator) Configure(conf map[string]string) {
	if val, ok := conf["concurrency"]; ok {
		conv, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			e.concurrency = int(conv)
		}
	}
}

func (e *JacobbiEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency": "max num of threads used (int)",
	}
}

func (e *JacobbiEstimator) PopulationPolicy(policy DatasetSimilarityPopulationPolicy) {
	e.popPolicy = policy
}

func (e *JacobbiEstimator) Serialize() []byte {
	buffer := new(bytes.Buffer)
	buffer.Write(getBytesInt(int(SIMILARITY_TYPE_JACOBBI)))
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

func (e *JacobbiEstimator) Deserialize(b []byte) {
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
	e.similarities = new(DatasetSimilarities)
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

// calculates a table line
func (e *JacobbiEstimator) calculateLine(start, lineNo int) {
	a := e.datasets[lineNo]
	for i := start; i < len(e.datasets); i++ {
		b := e.datasets[i]
		inter := len(DatasetsIntersection(a, b))
		union := len(DatasetsUnion(a, b))
		value := float64(inter) / float64(union)
		e.similarities.Set(lineNo, i, value)
	}
}
