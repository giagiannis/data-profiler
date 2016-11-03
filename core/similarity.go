package core

import (
	"errors"
	"log"
)

type DatasetSimilarityEstimator interface {
	Compute() error                     // computes the similarity matrix
	GetSimilarity(a, b Dataset) float64 // returns the similarity
}

// Factory method for creating a DatasetSimilarityEstimator
func NewDatasetSimilarityEstimator(
	estimatorType string,
	datasets []Dataset) DatasetSimilarityEstimator {
	if estimatorType == "jacobbi" {
		a := new(JacobbiEstimator)
		a.datasets = datasets
		a.concurrency = 8
		return a
	}
	return nil
}

// JacobbiEstimator estimates the Jacobbi coefficients between the different
// datasets. The Jacobbi coefficient between two datasets is defined as
// the cardinality of the intersection divided by the cardinality of the
// union of the two datasets.
type JacobbiEstimator struct {
	datasets     []Dataset      // the slice of datasets
	coefficients [][]float64    // holds the coefficients
	inverseIndex map[string]int // inverse index of datasets slice
	concurrency  int            // max threads running in parallel
}

func (e *JacobbiEstimator) Compute() error {
	log.Println("Calculating the inverse dataset index")
	if e.datasets == nil || len(e.datasets) == 0 {
		log.Println("No datasets were given")
		return errors.New("Empty dataset slice")
	}
	e.inverseIndex = make(map[string]int)
	for i := 0; i < len(e.datasets); i++ {
		e.inverseIndex[e.datasets[i].Id()] = i
	}

	log.Println("Fetching datasets in memory")
	for _, d := range e.datasets {
		d.ReadFromFile()
	}

	log.Println("Starting Jacobbi computation (parallel)")
	e.coefficients = make([][]float64, len(e.datasets)-1)
	for i := 0; i < len(e.datasets)-1; i++ {
		e.coefficients[i] = make([]float64, len(e.datasets)-i-1)
	}
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

func (e *JacobbiEstimator) GetSimilarity(a, b Dataset) float64 {
	idxA := e.inverseIndex[a.Id()]
	idxB := e.inverseIndex[b.Id()]
	if idxA == idxB {
		return 1.0
	} else if idxA > idxB {
		t := idxB
		idxB = idxA
		idxA = t
	}
	return e.coefficients[idxA][idxB-idxA-1]
}

// calculates a table line
func (e *JacobbiEstimator) calculateLine(lineNo int) {
	a := e.datasets[lineNo]
	for i := lineNo + 1; i < len(e.datasets); i++ {
		b := e.datasets[i]
		inter := len(DatasetsIntersection(&a, &b))
		union := len(DatasetsUnion(&a, &b))
		e.coefficients[lineNo][i-1-lineNo] = float64(inter) / float64(union)
	}
}
