package core

import (
	"errors"
	"log"
	"strconv"
)

// JacobbiEstimator estimates the Jacobbi coefficients between the different
// datasets. The Jacobbi coefficient between two datasets is defined as
// the cardinality of the intersection divided by the cardinality of the
// union of the two datasets.
type JacobbiEstimator struct {
	datasets     []*Dataset           // the slice of datasets
	similarities *DatasetSimilarities // holds the similarities
	concurrency  int                  // max threads running in parallel
}

func (e *JacobbiEstimator) Compute() error {
	e.similarities = NewDatasetSimilarities(e.datasets)
	log.Println("Fetching datasets in memory")
	if e.datasets == nil || len(e.datasets) == 0 {
		log.Println("No datasets were given")
		return errors.New("Empty dataset slice")
	}
	for _, d := range e.datasets {
		d.ReadFromFile()
	}

	log.Printf("Starting Jacobbi computation (%d threads)", e.concurrency)
	c := make(chan bool, e.concurrency)
	done := make(chan bool)
	for j := 0; j < e.concurrency; j++ {
		c <- true
	}
	for i := 0; i < len(e.datasets)-1; i++ {
		go func(c, done chan bool, i int) {
			<-c
			e.calculateLine(i)
			c <- true
			done <- true
		}(c, done, i)
	}
	for j := 0; j < len(e.datasets)-1; j++ {
		<-done
	}
	log.Println("Done")
	return nil
}

func (e *JacobbiEstimator) GetSimilarities() *DatasetSimilarities {
	return e.similarities
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

// calculates a table line
func (e *JacobbiEstimator) calculateLine(lineNo int) {
	a := e.datasets[lineNo]
	for i := lineNo + 1; i < len(e.datasets); i++ {
		b := e.datasets[i]
		inter := len(DatasetsIntersection(a, b))
		union := len(DatasetsUnion(a, b))
		value := float64(inter) / float64(union)
		e.similarities.Set(a.Path(), b.Path(), value)
	}
}
