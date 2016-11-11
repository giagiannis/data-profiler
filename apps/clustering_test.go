package apps

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/giagiannis/data-profiler/core"
)

func TestClustering(t *testing.T) {
	datasets := make([]*core.Dataset, 10)
	for i := 0; i < len(datasets); i++ {
		datasets[i] = core.NewDataset(fmt.Sprintf("data-%d", i))
	}
	sim := core.NewDatasetSimilarities(datasets)
	for _, d1 := range datasets {
		for _, d2 := range datasets {
			sim.Set(d1.Path(), d2.Path(), rand.Float64())
		}
	}

	cluster := NewClustering(sim)
	err := cluster.Compute()
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	fmt.Println(cluster.Results())
}

func TestNewDendrogram(t *testing.T) {
	datasets := make([]*core.Dataset, 40)
	for i := 0; i < len(datasets); i++ {
		datasets[i] = core.NewDataset(fmt.Sprintf("data-%d", i))
	}
	d := NewDendrogram(datasets)
	for unm := d.GetUnmerged(); d.root == nil; unm = d.GetUnmerged() {
		err := d.Merge(unm[0], unm[1])
		if err != nil {
			log.Println(err)
			t.FailNow()
		}
	}
	if len(d.unmerged) > 0 {
		log.Println("We still have unmerged nodes")
		t.FailNow()
	}
}
