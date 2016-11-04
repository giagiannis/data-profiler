package core

type BhattacharyyaEstimator struct {
	datasets     []Dataset            // datasets slice
	concurrency  int                  // the max number of threads that run in parallel
	similarities *DatasetSimilarities // the similarities struct
}

func (e *BhattacharyyaEstimator) Compute() error {
	// allocation of similarities struct
	e.similarities = NewDatasetSimilarities(e.datasets)
	return nil
}

func (e *BhattacharyyaEstimator) GetSimilarities() *DatasetSimilarities {
	return e.similarities
}
