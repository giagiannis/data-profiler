package core

import (
	"log"
	"math"
	"sort"
)

// Manager is responsible for synchronizing and managing the the
// analysis tasks. It also exports the coordinates of each dataset file to
// to the rest of the packages
type AnalyzerManager struct {
	datasets       []*Dataset        // dataset slice
	concurrency    int               // number of concurrent threads for the analysis
	analysisScript string            // script to execute for analysis
	results        map[string]Result // contains the results for the datasets
}

// NewManager is a constructor for the Manager class.
func NewManager(datasets []*Dataset, concurrency int, analysisScript string) *AnalyzerManager {
	m := AnalyzerManager{datasets, concurrency, analysisScript, nil}
	return &m
}

// Concurrency getter for Manager instance
func (m *AnalyzerManager) Concurrency() int {
	return m.concurrency
}

// Function used to begin the in-parallel analysis of the different datasets.
// The concurrency factor is set during the manager initialization
func (m *AnalyzerManager) Analyze() {
	analyzers := make([]ScriptAnalyzer, len(m.datasets))
	for i, d := range m.datasets {
		analyzers[i] = *NewScriptAnalyzer(*d, m.analysisScript)
		analyzers[i].script = m.analysisScript
	}
	log.Printf("Starting analysis (thread pool size: %d)\n", m.concurrency)

	//the finished threads post here
	done := make(chan bool)
	// represent the available exec slots
	cookies := make(chan bool, m.concurrency+1)

	for i := 0; i < m.concurrency; i++ {
		cookies <- true
	}
	for i, _ := range analyzers {
		// in parallel analysis of the datasets
		go func(i int, done chan bool, cookie chan bool) {
			<-cookie
			log.Printf("\tAnalysis thread %d started\n", i)
			exitStatus := analyzers[i].Analyze()
			if !exitStatus {
				log.Printf("\tERROR: Analysis thread %d failed\n", i)
			}
			cookie <- true
			done <- true
		}(i, done, cookies)
	}

	// wait for everyone to finish
	for i := 0; i < len(analyzers); i++ {
		<-done
	}
	log.Printf("Analysis finished")

	// write results here
	m.results = make(map[string]Result)
	avg := 0.0
	for _, v := range analyzers {
		avg += v.Duration()
		m.results[v.Dataset().Id()] = v.Result()
	}

}

// Results method is getter  for the results object
func (m *AnalyzerManager) Results() map[string]Result {
	return m.results
}

// OptimizationResultsPruning is used to prune the output dimension of each
// dataset according to the vector's energy (or information). Returns true
// if pruning was succesful.
func (m *AnalyzerManager) OptimizationResultsPruning() bool {
	var energy []float64
	for _, v := range m.results {
		if energy == nil {
			energy = make([]float64, len(v))
		}
		for i, k := range v {
			energy[i] += math.Pow(k, 2)
		}
	}
	kv := NewDimensionEnergyCollection(energy, 0.9)
	sort.Sort(sort.Reverse(kv))
	indices := kv.Cutoff()

	newResults := make(map[string]Result, len(m.results))
	for d, r := range m.results {
		newResults[d] = make([]float64, len(indices))
		for i, v := range indices {
			newResults[d][i] = r[v]
		}
	}
	m.results = newResults

	return true
}