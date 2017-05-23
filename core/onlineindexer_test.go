package core

import (
	"math/rand"
	"testing"
)

func TestNewOnlineIndexer(t *testing.T) {
	datasets := createPoolBasedDatasets(200, 100, 4)
	//estim := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACCARD, datasets)
	estim := NewDatasetSimilarityEstimator(SimilarityTypeBhattacharyya, datasets)
	estim.Configure(map[string]string{
		"concurrency": "10",
	})
	estim.Compute()
	md := NewMDScaling(estim.SimilarityMatrix(), 2, mdsScript)
	md.Compute()

	indexer := NewOnlineIndexer(estim, md.Coordinates(), onlineIndexingScript)
	coords, _, err := indexer.Calculate(datasets[rand.Int()%len(datasets)])
	if err != nil || coords == nil {
		t.Log(err)
		t.FailNow()
	}
	cleanDatasets(datasets)
}
