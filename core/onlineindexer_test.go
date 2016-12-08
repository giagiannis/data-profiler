package core

import (
	"math/rand"
	"os"
	"testing"
	"time"
)

const ONLINE_INDEXING_SCRIPT = "../_rscripts/sa.R"

func TestNewOnlineIndexer(t *testing.T) {
	rand.Seed(int64(time.Now().Nanosecond()))
	datasets := createPoolBasedDatasets(200, 100, 4)
	//estim := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACOBBI, datasets)
	estim := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_BHATTACHARYYA, datasets)
	estim.Configure(map[string]string{
		"concurrency": "10",
	})
	estim.Compute()
	md := NewMDScaling(estim.GetSimilarities(), 2, MDSCALING_SCRIPT)
	md.Compute()

	indexer := NewOnlineIndexer(estim, md.Coordinates(), ONLINE_INDEXING_SCRIPT)
	coords, stress, err := indexer.Calculate(datasets[rand.Int()%len(datasets)])
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(coords, stress)
	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
