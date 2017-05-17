package core

// CompositeSimilarityEstimator returns a similarity function based on compositions
// of simpler similarity expressions
type CompositeSimilarityEstimator struct {
	// the slice of datasets
	datasets []*Dataset
	// max threads running in parallel
	concurrency int
	// the matrix population policy
	popPolicy DatasetSimilarityPopulationPolicy
	// duration of the  execution
	duration float64
	// holds the similarities
	similarities *DatasetSimilarities

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
func (e *CompositeSimilarityEstimator) GetSimilarities() *DatasetSimilarities {
	// TODO: implement the method
	return nil
}
func (e *CompositeSimilarityEstimator) Configure(conf map[string]string) error {
	// TODO: implement the method
	return nil
}
func (e *CompositeSimilarityEstimator) Options() map[string]string {
	// TODO: implement the method
	return nil
}
func (e *CompositeSimilarityEstimator) PopulationPolicy(pol DatasetSimilarityPopulationPolicy) {
	// TODO: implement the method
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
