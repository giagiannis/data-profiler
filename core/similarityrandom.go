package core

import (
	"errors"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type RandomSimilarityEstimator struct {
	datasets     []*Dataset           // datasets slice
	similarities *DatasetSimilarities // similarities
	concurrency  int                  // max threads executed concurrently
}

func (e *RandomSimilarityEstimator) Compute() error {
	rand.Seed(int64(time.Now().Nanosecond()))
	e.similarities = NewDatasetSimilarities(e.datasets)
	log.Println("Fetching datasets in memory")
	if e.datasets == nil || len(e.datasets) == 0 {
		log.Println("No datasets were given")
		return errors.New("Empty dataset slice")
	}
	for _, d := range e.datasets {
		d.ReadFromFile()
	}

	log.Printf("Starting Random computation (%d threads)", e.concurrency)
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

func (e *RandomSimilarityEstimator) GetSimilarities() *DatasetSimilarities {
	return e.similarities
}

func (e *RandomSimilarityEstimator) Configure(conf map[string]string) {
	if val, ok := conf["concurrency"]; ok {
		conv, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			e.concurrency = int(conv)
		}
	}
}

func (e *RandomSimilarityEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency": "max num of threads used (int)",
	}
}

func (e *RandomSimilarityEstimator) calculateLine(lineNo int) {
	a := e.datasets[lineNo]
	for i := lineNo + 1; i < len(e.datasets); i++ {
		b := e.datasets[i]
		e.similarities.Set(a.Path(), b.Path(), rand.Float64())
	}
}
