package core

// Class used to execute online indexing. The user can supply a map containing
// distances from original datasets and the indexer returns the coordinates of the
// specified dataset.
type OnlineIndexer struct {
	coordinates []DatasetCoordinates       // the coordinates of the datasets
	estimator   DatasetSimilarityEstimator // estimator object to calculate distances
	script      string                     // script to evaluate the acceptable solutions

	datasetMap map[string]int // map used internally by the estimator to quickly translate paths to IDs
}

// NewOnlineIndexer is a constructor function used to initialize an OnlineIndexer
// object.
func NewOnlineIndexer(estimator DatasetSimilarityEstimator,
	coordinates []DatasetCoordinates,
	script string) *OnlineIndexer {
	indexer := new(OnlineIndexer)
	indexer.coordinates = coordinates
	indexer.estimator = estimator
	indexer.script = script
	return indexer
}

// Calculate method is responsible to calculate the coordinates of the specified
// dataset. In case that such a dataset cannot be represented by the specified
// coordinates system, an error is returned.
func (o *OnlineIndexer) Calculate(dataset *Dataset) (DatasetCoordinates, error) {
	// calculate the distances for the new dataset
	// generate file containing coordinates + distances

	//o.solveQuadSystem()
	return nil, nil
}

// solveQuadSystem solves the quadratic polynomial system in order to identify
// the coordinates of the new point.
func (o *OnlineIndexer) solveQuadSystem(fileName string) {
	// execute script

	// parse results
}
