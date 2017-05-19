package core

// CompositeSimilarityEstimator returns a similarity function based on compositions
// of simpler similarity expressions
type CompositeSimilarityEstimator struct {
	AbstractDatasetSimilarityEstimator

	// the slice of estimators to use for the similarity evaluation
	estimators []*DatasetSimilarityEstimator
}

func (e *CompositeSimilarityEstimator) Compute() error {
	// TODO: implement the method
	return nil
}

func (e *CompositeSimilarityEstimator) Similarity(a, b *Dataset) float64 {
	// TODO: implement the method
	return .0
}

func (e *CompositeSimilarityEstimator) Configure(conf map[string]string) error {
	// TODO: implement the method
	return nil
}

func (e *CompositeSimilarityEstimator) Options() map[string]string {
	// TODO: implement the method
	return nil
}

func (e *CompositeSimilarityEstimator) Serialize() []byte {
	// TODO: implement the method
	return nil
}

func (e *CompositeSimilarityEstimator) Deserialize([]byte) {
	// TODO: implement the method
}

func (e *CompositeSimilarityEstimator) Datasets() []*Dataset {
	return e.datasets
}

func (e *CompositeSimilarityEstimator) Duration() float64 {
	return e.duration
}
