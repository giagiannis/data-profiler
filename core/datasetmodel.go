package core

// Modeler is the interface for the objects that model the dataset space.
type Modeler interface {
	//  Configure is responsible to provide the necessary configuration
	// options to the modeler struct. Call it before Run.
	Configure(map[string]string)
	// Run initiates the modeling process.
	Run()

	// Datasets returns the datasets slice
	Datasets() []*Dataset
	// Samples returns the indices of the chosen datasets.
	Samples()
	// Actual returns a slice of the actual values
	RealValues() []float64
	// Approximated returns a slice of the approximated values
	AppxValues() []float64
}

// AbstractModeler implements the common methods of the Modeler structs
type AbstractModeler struct {
	datasets    []*Dataset
	evaluator   DatasetEvaluator
	coordinates []DatasetCoordinates

	samples    []int
	realValues []float64
	appxValues []float64
}

func (a *AbstractModeler) Configure(conf map[string]string) {}
func (a *AbstractModeler) Run() {
	// prepare training set
	// prepare test set
}

// Datasets returns the datasets slice
func (a *AbstractModeler) Datasets() []*Dataset {
	return a.datasets
}

// Samples return the indices of the chosen datasets
func (a *AbstractModeler) Samples() []int {
	return a.samples
}

// RealValues returns the actual values of the chosen datasets
func (a *AbstractModeler) RealValues() []float64 {
	return a.realValues
}

// AppxValues returns the values of all the datasets
func (a *AbstractModeler) AppxValues() []float64 {
	return a.appxValues
}
