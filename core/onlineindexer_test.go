package core

import (
	"math/rand"
	"os"
	"testing"
)

const ONLINE_INDEXING_SCRIPT = "../_rscripts/sa.R"

func TestNewOnlineIndexer(t *testing.T) {
	datasets := createPoolBasedDatasets(200, 100, 4)
	//estim := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACCARD, datasets)
	estim := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_BHATTACHARYYA, datasets)
	estim.Configure(map[string]string{
		"concurrency": "10",
	})
	estim.Compute()
	md := NewMDScaling(estim.GetSimilarities(), 2, MDSCALING_SCRIPT)
	md.Compute()

	indexer := NewOnlineIndexer(estim, md.Coordinates(), ONLINE_INDEXING_SCRIPT)
	coords, _, err := indexer.Calculate(datasets[rand.Int()%len(datasets)])
	if err != nil || coords == nil {
		t.Log(err)
		t.FailNow()
	}
	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
