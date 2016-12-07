package core

import (
	"math/rand"
	"os"
	"testing"
	"time"
)

const ONLINE_INDEXING_SCRIPT = "../r_scripts/quadsystem.R"

func TestNewOnlineIndexer(t *testing.T) {
	rand.Seed(int64(time.Now().Nanosecond()))
	datasets := createPoolBasedDatasets(2000, 50, 4)
	estim := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACOBBI, datasets[2:len(datasets)])
	estim.Configure(map[string]string{
		"concurrency": "10",
	})
	estim.Compute()
	md := NewMDScaling(estim.GetSimilarities(), 2, MDSCALING_SCRIPT)
	md.Compute()

	indexer := NewOnlineIndexer(estim, md.Coordinates(), ONLINE_INDEXING_SCRIPT)
	indexer.Calculate(datasets[1])
	indexer.Calculate(datasets[1])
	indexer.Calculate(datasets[1])
	indexer.Calculate(datasets[1])
	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
