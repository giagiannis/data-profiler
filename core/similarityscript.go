package core

import (
	"bytes"
	"errors"
	"log"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Script estimator utilizes a script to analyze the data based on some external
// algorithm and utilizes various norms to measure the differences between the
// analysis outputs.
type ScriptSimilarityEstimator struct {
	AbstractDatasetSimilarityEstimator
	analysisScript     string                        // the analysis script to be executed
	simType            ScriptSimilarityEstimatorType // similarity type - cosine, manhattan, euclidean
	inverseIndex       map[string]int                // inverse index that maps datasets to ints
	datasetCoordinates [][]float64                   // holds the dataset coordinates
}

type ScriptSimilarityEstimatorType uint8

const (
	SCRIPT_SIMILARITY_TYPE_MANHATTAN ScriptSimilarityEstimatorType = iota
	SCRIPT_SIMILARITY_TYPE_EUCLIDEAN ScriptSimilarityEstimatorType = iota + 1
	SCRIPT_SIMILARITY_TYPE_COSINE    ScriptSimilarityEstimatorType = iota + 2
)

func (s *ScriptSimilarityEstimator) Compute() error {
	if s.analysisScript == "" {
		log.Println("Analysis script not defined - exiting")
		return errors.New("Analysis script not defined")
	}

	// execute analysis for each dataset
	log.Println("Analyzing datasets")
	s.datasetCoordinates = s.analyzeDatasets()

	start := time.Now()
	// compare the analysis outcomes
	log.Println("Calculating similarities")
	s.similarities = NewDatasetSimilarities(len(s.datasets))
	s.inverseIndex = make(map[string]int)
	for i, d := range s.datasets {
		s.inverseIndex[d.Path()] = i
	}
	EstimatorCompute(s)
	s.duration = time.Since(start).Seconds()
	return nil
}
func (e *ScriptSimilarityEstimator) Duration() float64 {
	return e.duration
}

func (e *ScriptSimilarityEstimator) Similarity(a, b *Dataset) float64 {
	var coordsA, coordsB []float64
	if id, ok := e.inverseIndex[a.Path()]; ok {
		coordsA = e.datasetCoordinates[id]
	} else {
		coordsA = e.analyzeDataset(a.Path())
	}
	if id, ok := e.inverseIndex[b.Path()]; ok {
		coordsB = e.datasetCoordinates[id]
	} else {
		coordsB = e.analyzeDataset(b.Path())
	}
	if e.simType == SCRIPT_SIMILARITY_TYPE_COSINE {
		val, err := e.cosine(coordsA, coordsB)
		if err != nil {
			log.Println(err)
		}
		return val
	}
	normDegree := 2 // default is EUCLIDEAN distance
	if e.simType == SCRIPT_SIMILARITY_TYPE_MANHATTAN {
		normDegree = 1
	}
	val, err := e.norm(coordsA, coordsB, normDegree)
	if err != nil {
		log.Println(err)
	}
	return DistanceToSimilarity(val)
}

func (e *ScriptSimilarityEstimator) Datasets() []*Dataset {
	return e.datasets
}

func (s *ScriptSimilarityEstimator) SimilarityMatrix() *DatasetSimilarityMatrix {
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
	if val, ok := conf["type"]; ok {
		if val == "cosine" {
			s.simType = SCRIPT_SIMILARITY_TYPE_COSINE
		} else if val == "manhattan" {
			s.simType = SCRIPT_SIMILARITY_TYPE_MANHATTAN
		} else if val == "euclidean" {
			s.simType = SCRIPT_SIMILARITY_TYPE_EUCLIDEAN
		} else {
			log.Println("Similarity Type not known, valid values: [cosine manhattan euclidean]")
		}
	}

}

func (s *ScriptSimilarityEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency": "max number of threads to run in parallel",
		"script":      "path of the analysis script to be executed",
		"type":        "the type of the similarity - one of:  [cosine manhattan euclidean]",
	}
}

func (e *ScriptSimilarityEstimator) SetPopulationPolicy(policy DatasetSimilarityPopulationPolicy) {
	e.popPolicy = policy
}

func (e *ScriptSimilarityEstimator) Serialize() []byte {
	buffer := new(bytes.Buffer)
	buffer.Write(getBytesInt(int(SIMILARITY_TYPE_SCRIPT)))
	buffer.Write(getBytesInt(e.concurrency))
	buffer.Write(getBytesInt(int(e.simType)))
	buffer.WriteString(e.analysisScript + "\n")

	pol := e.popPolicy.Serialize()
	log.Println(len(pol))
	buffer.Write(getBytesInt(len(pol)))
	buffer.Write(pol)

	sim := e.similarities.Serialize()
	log.Println(len(sim))
	buffer.Write(getBytesInt(len(sim)))
	buffer.Write(sim)

	// serialize dataset names
	buffer.Write(getBytesInt(len(e.datasets)))
	for _, d := range e.datasets {
		buffer.WriteString(d.Path() + "\n")
	}

	// write number of coordinates per dataset
	buffer.Write(getBytesInt(len(e.datasetCoordinates[0])))
	for _, arr := range e.datasetCoordinates {
		for _, v := range arr {
			buffer.Write(getBytesFloat(v))
		}
	}
	return buffer.Bytes()
}

func (e *ScriptSimilarityEstimator) Deserialize(b []byte) {
	buffer := bytes.NewBuffer(b)
	tempInit := make([]byte, 4)
	buffer.Read(tempInit) // consume estimator type
	var count int
	buffer.Read(tempInit)
	e.concurrency = getIntBytes(tempInit)
	buffer.Read(tempInit)
	e.simType = ScriptSimilarityEstimatorType(getIntBytes(tempInit))
	line, _ := buffer.ReadString('\n')
	e.analysisScript = strings.TrimSpace(line)

	buffer.Read(tempInit)
	count = getIntBytes(tempInit)
	polBytes := make([]byte, count)
	buffer.Read(polBytes)
	e.popPolicy = *new(DatasetSimilarityPopulationPolicy)
	e.popPolicy.Deserialize(polBytes)

	buffer.Read(tempInit)
	count = getIntBytes(tempInit)
	similarityBytes := make([]byte, count)
	buffer.Read(similarityBytes)
	e.similarities = new(DatasetSimilarityMatrix)
	e.similarities.Deserialize(similarityBytes)

	buffer.Read(tempInit)
	count = getIntBytes(tempInit)
	e.datasets = make([]*Dataset, count)
	e.inverseIndex = make(map[string]int)
	for i := range e.datasets {
		line, _ := buffer.ReadString('\n')
		line = strings.TrimSpace(line)
		e.datasets[i] = NewDataset(line)
		e.inverseIndex[line] = i
	}

	tempFloat := make([]byte, 8)
	buffer.Read(tempInit)
	count = getIntBytes(tempInit)
	e.datasetCoordinates = make([][]float64, len(e.datasets))
	for i := range e.datasets {
		e.datasetCoordinates[i] = make([]float64, count)
		for j := range e.datasetCoordinates[i] {
			buffer.Read(tempFloat)
			e.datasetCoordinates[i][j] = getFloatBytes(tempFloat)
		}
	}
}

func (s *ScriptSimilarityEstimator) analyzeDatasets() [][]float64 {
	c, done := make(chan bool, s.concurrency), make(chan bool)
	coords := make([][]float64, len(s.datasets))
	for i := 0; i < s.concurrency; i++ {
		c <- true
	}
	for i, d := range s.datasets {
		go func(c, done chan bool, i int, path string) {
			<-c
			coords[i] = s.analyzeDataset(path)
			c <- true
			done <- true
		}(c, done, i, d.Path())
	}

	for i := 0; i < len(s.datasets); i++ {
		<-done
	}
	return coords
}

// analyzeDataset executed the analysis script into the specified dataset
func (s *ScriptSimilarityEstimator) analyzeDataset(path string) []float64 {
	log.Println("Analyzing", path)
	cmd := exec.Command(s.analysisScript, path)
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
	}
	results := make([]float64, 0)
	for _, sv := range strings.Split(string(out), "\t") {
		conv, err := strconv.ParseFloat(strings.TrimSpace((sv)), 64)
		if err == nil {
			results = append(results, conv)
		} else {
			log.Println(err)
		}
	}
	log.Println("Tuple read:", results)
	return results
}

// norm function calculates the norm between two float slices
func (s *ScriptSimilarityEstimator) norm(a, b []float64, normDegree int) (float64, error) {
	if len(a) != len(b) {
		return -1, errors.New("arrays have different sizes")
	}
	sum := 0.0
	for i := range a {
		dif := math.Abs(a[i] - b[i])
		sum += math.Pow(dif, float64(normDegree))
	}
	return math.Pow(sum, 1.0/float64(normDegree)), nil
}

// cosine calculates the cosine similarity between two vectors
func (s *ScriptSimilarityEstimator) cosine(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return -1, errors.New("arrays have different sizes")
	}
	nomin, sumA, sumB := 0.0, 0.0, 0.0
	for i := range a {
		nomin += a[i] * b[i]
		sumA += a[i] * a[i]
		sumB += b[i] * b[i]
	}
	denom := math.Sqrt(sumA) * math.Sqrt(sumB)
	if denom == 0.0 {
		return -1, errors.New("Zero denominator to cosine similarity")
	}
	return nomin / denom, nil
}
