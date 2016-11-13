package apps

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/giagiannis/data-profiler/core"
)

func TestClustering(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	datasets := make([]*core.Dataset, 500)
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
	cluster.SetConcurrency(10)
	err := cluster.Compute()
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	max, min := cluster.Results().Heights()
	fmt.Println(max, min)
	for i := 0; i <= max+10; i++ {
		sum := 0
		for _, cl := range cluster.Results().GetClusters(i) {
			sum += len(cl)
		}
		if sum != len(datasets) {
			log.Println("Datasets lost in clusters")
			t.FailNow()
		}
	}
	unary := false
	for _, cl := range cluster.Results().GetClusters(min) {
		unary = unary || (len(cl) == 1)
	}
	if !unary {
		log.Println("No unary cluster found for min height")
		t.FailNow()
	}
	unary = true
	for _, cl := range cluster.Results().GetClusters(max) {
		unary = unary && (len(cl) == 1)
	}
	if !unary {
		log.Println("Cluster with more than 1 datasets found, although at max height")
		t.FailNow()
	}

}

func TestNewDendrogram(t *testing.T) {
	datasets := make([]*core.Dataset, 40)
	for i := 0; i < len(datasets); i++ {
		datasets[i] = core.NewDataset(fmt.Sprintf("data-%d", i))
	}
	d := NewDendrogram(datasets)
	for unm := d.getUnmerged(); d.root == nil; unm = d.getUnmerged() {
		err := d.merge(unm[0], unm[1])
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
