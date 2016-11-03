package optimization

import (
	"log"
	"math"
	"math/rand"

	"github.com/giagiannis/data-profiler/core"
)

// SimulatedAnnealingOptimizer executes the Simaluted Annealing optimization
// algorithm
type SimulatedAnnealingOptimizer struct {
	OptimizerBase                        // Anonymous field, used to extend OptimizerBase
	maxIterations int                    // max iterations of SA
	tempDecay     float64                // decay of the temperature factor
	tempInit      float64                // initial temperature (of first iteration
	coefficients  map[string]core.Result // coefficients of the datasets
	datasets      []core.Dataset         // list of datasets
	distanceType  core.DistanceType      // type of distance to use
}

// NewSimulatedAnnealingOptimizer is  the default constructor used to allocate
// a new SimulatedAnnealingOptimizer instance.
func NewSimulatedAnnealingOptimizer(
	scriptName string,
	testDataset core.Dataset,
	maxIterations int,
	tempDecay float64,
	tempInit float64,
	datasets []core.Dataset,
	coefficients map[string]core.Result,
	distanceType core.DistanceType) *SimulatedAnnealingOptimizer {

	o := new(SimulatedAnnealingOptimizer)
	o.OptimizerBase = *new(OptimizerBase)
	o.OptimizerBase.execScript = scriptName
	o.OptimizerBase.testDataset = testDataset
	o.maxIterations = maxIterations
	o.tempDecay = tempDecay
	o.tempInit = tempInit
	o.coefficients = coefficients
	o.datasets = datasets
	o.distanceType = distanceType
	return o
}

// Run executes the optimizer
func (o *SimulatedAnnealingOptimizer) Run() {
	log.Println("Optimizer execution started")
	currentState := o.initialDataset()
	currentScore, _ := o.Execute(currentState)
	log.Printf("Initial dataset picked: (%s, %.5f)\n", currentState, currentScore)
	for i := 0; i < o.maxIterations; i++ {
		t := o.temperature(i)
		candidateState := o.neighbor(currentState, t)
		candidateScore, _ := o.Execute(candidateState)
		log.Printf("\t(it %d) - candidate: (%s, %.5f)\n", i+1, candidateState, candidateScore)

		if o.acceptance(currentScore, candidateScore, t) {
			currentState = candidateState
			currentScore = candidateScore
		}
		log.Printf("\t(it %d) - picked:    (%s, %.5f)\n", i+1, currentState, currentScore)
	}
	o.result = OptimizerResult{currentState, currentScore}
	log.Println("Optimizer finished")
}

// Returns a random initial dataset
func (o *SimulatedAnnealingOptimizer) initialDataset() core.Dataset {
	randomIndex := rand.Int() % len(o.datasets)
	return o.datasets[randomIndex]
}

// Returns the temperature decay factor
func (o *SimulatedAnnealingOptimizer) temperature(currentIteration int) float64 {
	return o.tempInit * math.Pow(o.tempDecay, float64(currentIteration))
}

// Returns a neighbor of the current state - relevant to the temperature
func (o *SimulatedAnnealingOptimizer) neighbor(
	current core.Dataset,
	temperature float64) core.Dataset {
	probabilities := make([]float64, len(o.datasets))
	distances := make([]float64, len(o.datasets))
	sum := 0.0
	for i, d := range o.datasets {
		if d.Id() != current.Id() {
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
	return o.datasets[ind]
}

// datasetDistance is used to measure the distance between two datasets
func (o *SimulatedAnnealingOptimizer) datasetsDistance(d1 core.Dataset, d2 core.Dataset) float64 {
	v1, v2 := o.coefficients[d1.Id()], o.coefficients[d2.Id()]
	return core.Distance(v1, v2, o.distanceType)
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
