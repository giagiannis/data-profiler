package apps

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/giagiannis/data-profiler/core"
)

// Clustering struct is responsible to execute the necessary actions in order
// to cluster the datasets based on their availability
type Clustering struct {
	similarities *core.DatasetSimilarities // dataset similarities
	results      *Dendrogram               // holds the clustering results
}

// Constructor for creating a Clustering object, providing a DatasetSimilarities
// object
func NewClustering(similarities *core.DatasetSimilarities) *Clustering {
	c := new(Clustering)
	c.similarities = similarities
	return c
}

// Used to configure the clustering object
func (c *Clustering) Configure(map[string]string) {
	// does nothing for now
}

// Executes the clustering
func (c *Clustering) Compute() error {
	c.results = NewDendrogram(c.similarities.Datasets())
	for c.results.HasUnmerged() {
		unmerged := c.results.GetUnmerged()
		c.mergeClosestPairs(unmerged)
	}
	return nil
}

// Returns the results
func (c *Clustering) Results() *Dendrogram {
	return c.results
}

// Returns the similarity between two different clusters of datasets
func (c *Clustering) getClustersSimilarity(a, b []*core.Dataset) float64 {
	sum := 0.0
	for _, m1 := range a {
		for _, m2 := range b {
			sum += c.similarities.Get(m1.Path(), m2.Path())
		}
	}
	return sum / float64(len(a)*len(b))
}

// Evaluates and merges the most similar clusters at each turn
func (c *Clustering) mergeClosestPairs(unmerged []*DendrogramNode) {
	closestNode, maxSimilarity := make(map[int]int), make(map[int]float64)
	for i := 0; i < len(unmerged); i++ {
		for j := 0; j < len(unmerged); j++ {
			n1, n2 := unmerged[i], unmerged[j]
			val := c.getClustersSimilarity(n1.datasets, n2.datasets)
			old, ok := maxSimilarity[i]
			if i != j && (!ok || old < val) {
				maxSimilarity[i] = val
				closestNode[i] = j
			}
		}
	}
	for k, v := range closestNode {
		if closestNode[v] == k {
			c.results.Merge(unmerged[k], unmerged[v])
		}
	}

}

// Dendrogram represents the results of the ClusterApp objects
type Dendrogram struct {
	root     *DendrogramNode         // the root of the tree
	unmerged map[int]*DendrogramNode // Dendrogram leaf nodes, indexed by their id
}

type DendrogramNode struct {
	id          int
	datasets    []*core.Dataset
	father      *DendrogramNode
	left, right *DendrogramNode
}

func (n DendrogramNode) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, d := range n.datasets {
		buf.WriteString(fmt.Sprintf("%s", (*d).Path()))
		if i < len(n.datasets)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("}")
	return buf.String()
}

// Dendrogram constructor
func NewDendrogram(datasets []*core.Dataset) *Dendrogram {
	d := new(Dendrogram)
	d.root = nil
	d.unmerged = make(map[int]*DendrogramNode)
	for i := 0; i < len(datasets); i++ {
		d.unmerged[i] = &DendrogramNode{i,
			[]*core.Dataset{datasets[i]},
			nil, nil, nil}
	}

	return d
}

func (d *Dendrogram) HasUnmerged() bool {
	return d.root == nil
}

// Returns a list of the unmerged nodes.
func (d *Dendrogram) GetUnmerged() []*DendrogramNode {
	res := make([]*DendrogramNode, len(d.unmerged))
	i := 0
	for _, node := range d.unmerged {
		res[i] = node
		i++
	}
	return res
}

// Merge method is used to merge nodes of the tree that have not been merged yet
func (d *Dendrogram) Merge(a, b *DendrogramNode) error {
	if _, ok := d.unmerged[a.id]; !ok {
		return errors.New("Node already merged or not known")
	} else if _, ok := d.unmerged[b.id]; !ok {
		return errors.New("Node already merged or not known")
	}
	delete(d.unmerged, a.id)
	delete(d.unmerged, b.id)
	newNode := new(DendrogramNode)
	newNode.id = a.id
	newNode.left, newNode.right, newNode.father = a, b, nil
	a.father, b.father = newNode, newNode
	newNode.datasets = append(newNode.datasets, a.datasets...)
	newNode.datasets = append(newNode.datasets, b.datasets...)
	if len(d.unmerged) == 0 { // the new on is the root
		d.root = newNode
	} else {
		d.unmerged[newNode.id] = newNode
	}
	return nil
}

func (d *Dendrogram) String() string {
	spa := "\t"
	var buf bytes.Buffer
	var dfs func(*DendrogramNode, string)
	dfs = func(node *DendrogramNode, pad string) {
		buf.WriteString(fmt.Sprintf("%s%s\n", pad, node.String()))
		if node.left != nil && node.right != nil {
			dfs(node.left, pad+spa)
			dfs(node.right, pad+spa)
		}
	}
	dfs(d.root, "")
	return buf.String()
}
