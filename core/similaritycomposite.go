package core

// CompositeSimilarityEstimator returns a similarity function based on
// compositions of simpler similarity expressions. The user needs to provide a
// formula containing the expression of the similarity function, e.g.:
// 0.8 x BHATTACHARRYA + 0.2 CORRELATION
// Note that it is the user's responsibility to guarantee that the overall
// expression remains within the limits of the similarity expression [0,1].
type CompositeSimilarityEstimator struct {
	AbstractDatasetSimilarityEstimator

	// the slice of estimators to use for the similarity evaluation
	estimators []*DatasetSimilarityEstimator
	// the math expression used for the estimation
	expression string
}

// Compute runs the similarity estimation and forms the SimilarityMatrix
func (e *CompositeSimilarityEstimator) Compute() error {
	// TODO: implement the method
	return nil
}

// Similarity returns the similarity between two datasets
func (e *CompositeSimilarityEstimator) Similarity(a, b *Dataset) float64 {
	// TODO: implement the method
	return .0
}

// Configure provides the configuration parameters needed by the Estimator
func (e *CompositeSimilarityEstimator) Configure(conf map[string]string) error {
	// TODO: implement the method
	return nil
}

// Options returns the applicable parameters needed by the Estimator.
func (e *CompositeSimilarityEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency": "max number of threads to be run in parallel",
		"formula":     "the math expression that combines the estimators",
	}
}

// Serialize returns an array of bytes representing the Estimator.
func (e *CompositeSimilarityEstimator) Serialize() []byte {
	// TODO: implement the method
	return nil
}

// Deserialize constructs an Estimator object based on the byte array provided.
func (e *CompositeSimilarityEstimator) Deserialize([]byte) {
	// TODO: implement the method
}
