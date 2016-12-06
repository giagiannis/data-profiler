package core

import (
	"math/rand"
	"os"
	"testing"
)

const ONLINE_INDEXING_SCRIPT = "../r_scripts/quadsystem/quadsystem.R"

func TestNewOnlineIndexer(t *testing.T) {
	datasets := createPoolBasedDatasets(2000, 10, 4)
	estim := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACOBBI, datasets)
	estim.Compute()

	coords := make([]DatasetCoordinates, 10)
	for i := range coords {
		coords[i] = make([]float64, 2)
		for j := range coords[i] {
			coords[i][j] = rand.Float64()
		}
	}

	//indexer := NewOnlineIndexer(estim, coords, ONLINE_INDEXING_SCRIPT)
	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
