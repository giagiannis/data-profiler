package core

import (
	"errors"
	"log"
)

// JacobbiEstimator estimates the Jacobbi coefficients between the different
// datasets. The Jacobbi coefficient between two datasets is defined as
// the cardinality of the intersection divided by the cardinality of the
// union of the two datasets.
type JacobbiEstimator struct {
	datasets     []Dataset            // the slice of datasets
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

	log.Println("Starting Jacobbi computation (parallel)")
	c := make(chan bool, len(e.datasets)+1)
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
	for j := 0; j < 8; j++ {
		<-done
	}
	log.Println("Done")
	return nil
}

func (e *JacobbiEstimator) GetSimilarities() *DatasetSimilarities {
	return e.similarities
}

// calculates a table line
func (e *JacobbiEstimator) calculateLine(lineNo int) {
	a := e.datasets[lineNo]
	for i := lineNo + 1; i < len(e.datasets); i++ {
		b := e.datasets[i]
		inter := len(DatasetsIntersection(&a, &b))
		union := len(DatasetsUnion(&a, &b))
		value := float64(inter) / float64(union)
		e.similarities.Set(a, b, value)
	}
}
