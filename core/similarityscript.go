package core

import (
	"errors"
	"log"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// Script estimator utilizes a script to analyze the data based on some external
// algorithm and utilizes various norms to measure the differences between the
// analysis outputs.
type ScriptSimilarityEstimator struct {
	analysisScript string               // the analysis script to be executed
	datasets       []*Dataset           // the input datasets
	concurrency    int                  // max number of threads to run in parallel
	similarities   *DatasetSimilarities // the similarities struct
	normDegree     int                  // defines the degree of the norm
}

// Compute executes the similarity algorithm in order to provide
func (s *ScriptSimilarityEstimator) Compute() error {
	if s.analysisScript == "" {
		log.Println("Analysis script not defined - exiting")
		return errors.New("Analysis script not defined")
	}
	if s.normDegree < 1 {
		log.Println("Cannot estimate a norm with degree lower than one")
		return errors.New("Norm degree less than one")
	}

	// execute analysis for each dataset
	log.Println("Analyzing datasets")
	coordinates := s.analyzeDatasets()

	// compare the analysis outcomes
	log.Println("Calculating similarities")
	s.similarities = NewDatasetSimilarities(s.datasets)
	c, done := make(chan bool, s.concurrency), make(chan bool)
	for i := 0; i < s.concurrency; i++ {
		c <- true
	}

	for i := 0; i < len(s.datasets); i++ {
		go func(c, done chan bool, line int) {
			<-c
			s.computeLine(coordinates, line)
			c <- true
			done <- true
		}(c, done, i)
	}

	for i := 0; i < len(s.datasets); i++ {
		<-done
	}
	return nil
}

func (s *ScriptSimilarityEstimator) GetSimilarities() *DatasetSimilarities {
	return s.similarities
}

func (s *ScriptSimilarityEstimator) Configure(conf map[string]string) {
	if val, ok := conf["concurrency"]; ok {
		conv, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			s.concurrency = int(conv)
		}
	}
	if val, ok := conf["script"]; ok {
		s.analysisScript = val
	}
	if val, ok := conf["norm"]; ok {
		conv, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			s.normDegree = int(conv)
		}
	}

}

func (s *ScriptSimilarityEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency": "max number of threads to run in parallel",
		"script":      "path of the analysis script to be executed",
		"norm":        "the degree of the norm to be used among datasets",
	}
}

func (s *ScriptSimilarityEstimator) analyzeDatasets() [][]float64 {
	c, done := make(chan bool, s.concurrency), make(chan bool)
	coords := make([][]float64, len(s.datasets))
	for i := 0; i < s.concurrency; i++ {
		c <- true
	}
	for i, d := range s.datasets {
		go func(c, done chan bool, i int) {
			<-c
			coords[i] = s.analyzeDataset(d.Path())
			c <- true
			done <- true
		}(c, done, i)
	}

	for i := 0; i < len(s.datasets); i++ {
		<-done
	}
	return coords
}

// analyzeDataset executed the analysis script into the specified dataset
func (s *ScriptSimilarityEstimator) analyzeDataset(path string) []float64 {
	cmd := exec.Command(s.analysisScript, path)
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
	}
	results := make([]float64, 0)
	for _, sv := range strings.Split(string(out), " ") {
		conv, err := strconv.ParseFloat(sv, 64)
		if err == nil {
			results = append(results, conv)
		}
	}
	return results
}

// norm function calculates the norm between two float slices
func (s *ScriptSimilarityEstimator) norm(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return -1, errors.New("Arrays have different sizes")
	}
	sum := 0.0
	for i := range a {
		sum += math.Pow(a[i]-b[i], float64(s.normDegree))
	}
	return math.Pow(sum, 1.0/float64(s.normDegree)), nil
}

func (s *ScriptSimilarityEstimator) computeLine(coordinates [][]float64, line int) {
	a := s.datasets[line].Path()
	for j := line + 1; j < len(coordinates[line]); j++ {
		b := s.datasets[j].Path()
		v, err := s.norm(coordinates[line], coordinates[j])
		// converting distance to similarity
		sim := 1.0 / (1.0 + v)
		if err != nil {
			log.Panic(err)
		}
		s.similarities.Set(a, b, sim)
	}
}
