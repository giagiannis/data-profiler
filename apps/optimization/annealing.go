package optimization

import (
	"log"
	"math"
	"math/rand"

	"github.com/giagiannis/data-profiler/analysis"
)

// SimulatedAnnealingOptimizer executes the Simaluted Annealing optimization
// algorithm
type SimulatedAnnealingOptimizer struct {
	OptimizerBase                                      // Anonymous field, used to extend OptimizerBase
	maxIterations int                                  // max iterations of SA
	tempDecay     float64                              // decay of the temperature factor
	tempInit      float64                              // initial temperature (of first iteration
	coefficients  map[analysis.Dataset]analysis.Result // coefficients of the datasets
	datasets      []analysis.Dataset                   // list of datasets
	distanceType  analysis.DistanceType                // type of distance to use
}

// NewSimulatedAnnealingOptimizer is  the default constructor used to allocate
// a new SimulatedAnnealingOptimizer instance.
func NewSimulatedAnnealingOptimizer(
	scriptName string,
	testDataset analysis.Dataset,
	maxIterations int,
	tempDecay float64,
	tempInit float64,
	coefficients map[analysis.Dataset]analysis.Result,
	distanceType analysis.DistanceType) *SimulatedAnnealingOptimizer {

	o := new(SimulatedAnnealingOptimizer)
	o.OptimizerBase = *new(OptimizerBase)
	o.OptimizerBase.execScript = scriptName
	o.OptimizerBase.testDataset = testDataset
	o.maxIterations = maxIterations
	o.tempDecay = tempDecay
	o.tempInit = tempInit
	o.coefficients = coefficients
	o.datasets = make([]analysis.Dataset, len(coefficients))
	o.distanceType = distanceType
	i := 0
	for k := range o.coefficients {
		o.datasets[i] = k
		i++
	}
	return o
}

// Run executes the optimizer
func (o *SimulatedAnnealingOptimizer) Run() {
	log.Print("Optimizer execution started")
	log.Print("\tPicking initial dataset")
	currentState := o.initialDataset()
	log.Printf("\tDataset picked: (%s)\n", currentState)
	log.Printf("\tObtaining dataset value\n")
	currentScore, _ := o.Execute(currentState)
	log.Printf("\tValue obtained: %.5f\n", currentScore)
	for i := 0; i < o.maxIterations; i++ {
		t := o.temperature(i)
		candidateState := o.neighbor(currentState, t)
		candidateScore, _ := o.Execute(candidateState)
		log.Printf("\t\t(it %d) - candidate status: (%s, %.5f)\n", i, candidateState, candidateScore)

		if o.acceptance(currentScore, candidateScore, t) {
			currentState = candidateState
			currentScore = candidateScore
		}
		log.Printf("\t\t(it %d) - picked state: (%s, %.5f)\n", i, currentState, currentScore)
	}
	o.result = OptimizerResult{currentState, currentScore}
}

// Returns a random initial dataset
func (o *SimulatedAnnealingOptimizer) initialDataset() analysis.Dataset {
	randomIndex := rand.Int() % len(o.datasets)
	return o.datasets[randomIndex]
}

// Returns the temperature decay factor
func (o *SimulatedAnnealingOptimizer) temperature(currentIteration int) float64 {
	return o.tempInit * math.Pow(o.tempDecay, float64(currentIteration))
}

// Returns a neighbor of the current state - relevant to the temperature
func (o *SimulatedAnnealingOptimizer) neighbor(
	current analysis.Dataset,
	temperature float64) analysis.Dataset {
	probabilities := make([]float64, len(o.datasets))
	distances := make([]float64, len(o.datasets))
	sum := 0.0
	for i, d := range o.datasets {
		if d != current {
			distances[i] = o.datasetsDistance(current, d)
			probabilities[i] =
				o.probabilityDensityFunction(
					distances[i],
					temperature)
			sum += probabilities[i]
		}
	}
	for i := range probabilities {
		probabilities[i] /= sum
	}
	cummulative := make([]float64, len(probabilities))
	for i := range probabilities {
		if i > 0 {
			cummulative[i] = cummulative[i-1] + probabilities[i]
		} else {
			cummulative[i] = probabilities[i]
		}
	}
	ind :=
		func(val float64, cummulative []float64) int {
			start := 0
			stop := len(cummulative) - 1
			for {
				mid := (start + stop) / 2
				if mid == start || mid == stop {
					if cummulative[mid] < val {
						return mid + 1
					}
					return mid
				}
				if cummulative[mid] < val {
					start = mid + 1
				} else {
					stop = mid
				}

			}
			return len(cummulative) - 1
		}(rand.Float64(), cummulative)

		//	for i, v := range cummulative {
		//		if o.datasets[i] == current {
		//			log.Printf("\t\tC\t%s) %.5f\t[%.5f]\n", o.datasets[i].Id(), v, distances[i])
		//		} else if i == ind {
		//			log.Printf("\t\tc\t%s) %.5f\t[%.5f]\n", o.datasets[i].Id(), v, distances[i])
		//		} else {
		//			log.Printf("\t\t\t%s) %.5f\t[%.5f]\n", o.datasets[i].Id(), v, distances[i])
		//		}
		//	}

	return o.datasets[ind]
}

// datasetDistance is used to measure the distance between two datasets
func (o *SimulatedAnnealingOptimizer) datasetsDistance(d1 analysis.Dataset, d2 analysis.Dataset) float64 {
	v1, v2 := o.coefficients[d1], o.coefficients[d2]
	return analysis.Distance(v1, v2, o.distanceType)
}

// Represents the acceptance probability
func (o *SimulatedAnnealingOptimizer) acceptance(
	current float64,
	candidate float64,
	temperature float64) bool {
	delta := candidate - current
	threshold := o.probabilityDensityFunction(delta, temperature)
	randomNumber := rand.Float64()
	return (randomNumber < threshold)
}

// Util function used to calculate the the value 1/(1 + exp(x/y)). Used both
// as an acceptance PDF and as a neighborhood PDF
func (o *SimulatedAnnealingOptimizer) probabilityDensityFunction(x float64, y float64) float64 {
	return 1.0 / (1.0 + math.Exp(x/y))
}
