package core

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
)

type BhattacharyyaEstimator struct {
	datasets          []*Dataset           // datasets slice
	similarities      *DatasetSimilarities // the similarities struct
	concurrency       int                  // the max number of threads that run in parallel
	kdTreeScaleFactor float64              // determines the height of the kd tree to be used
}

func (e *BhattacharyyaEstimator) Compute() error {
	// allocation of similarities struct
	e.similarities = NewDatasetSimilarities(e.datasets)

	log.Println("Fetching datasets in memory")
	if e.datasets == nil || len(e.datasets) == 0 {
		log.Println("No datasets were given")
		return errors.New("Empty dataset slice")
	}
	for i, d := range e.datasets {
		log.Println(i, d.Path())
		d.ReadFromFile()
	}

	log.Println("Estimating a KD-tree partition")
	tree := NewKDTreePartition(e.datasets[0].Data())
	oldHeight := tree.Height()
	newHeight := int(float64(oldHeight) * e.kdTreeScaleFactor)
	log.Printf("Pruning the tree from height %d to %d\n", oldHeight, newHeight)
	tree.Prune(newHeight)
	//tree.Prune(1)
	indices := make(map[string][]int)
	for _, d := range e.datasets {
		indices[d.Path()] = tree.GetLeafIndex(d.Data())
	}

	log.Println("Computing the similarities using", e.concurrency, "threads")
	for i := 0; i < len(e.datasets); i++ {
		e.computeLine(i, indices)
	}
	c := make(chan bool, e.concurrency)
	done := make(chan bool)
	for j := 0; j < e.concurrency; j++ {
		c <- true
	}
	for i := 0; i < len(e.datasets)-1; i++ {
		go func(c, done chan bool, i int, indices map[string][]int) {
			<-c
			e.computeLine(i, indices)
			c <- true
			done <- true
		}(c, done, i, indices)
	}
	for j := 0; j < len(e.datasets)-1; j++ {
		<-done
	}

	log.Println("Done")
	return nil
}

func (e *BhattacharyyaEstimator) GetSimilarities() *DatasetSimilarities {
	return e.similarities
}

func (e *BhattacharyyaEstimator) Configure(conf map[string]string) {
	if val, ok := conf["concurrency"]; ok {
		conv, err := strconv.ParseInt(val, 10, 32)
		e.concurrency = int(conv)
		if err != nil {
			log.Println(err)
		}
	}
	if val, ok := conf["tree.scale"]; ok {
		//conv, err := strconv.ParseInt(val, 10, 32)
		conv, err := strconv.ParseFloat(val, 64)
		e.kdTreeScaleFactor = conv
		if err != nil {
			log.Println(err)
		}
	}
}

func (e *BhattacharyyaEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency": "max num of threads used (int)",
		"tree.scale":  "determines the portion of the kd-tree to be used",
	}
}

func (e *BhattacharyyaEstimator) computeLine(i int, indices map[string][]int) {
	for j := i + 1; j < len(e.datasets); j++ {
		a, b := e.datasets[i], e.datasets[j]
		r1, r2 := indices[a.Path()], indices[b.Path()]
		sum := 0.0
		for k := 0; k < len(r1); k++ {
			sum += math.Sqrt(float64(r1[k] * r2[k]))
		}
		sum /= math.Sqrt(float64(len(a.Data()) * len(b.Data())))
		e.similarities.Set(a.Path(), b.Path(), sum)
	}
}

type kdTreeNode struct {
	dim   int
	value float64
	right *kdTreeNode
	left  *kdTreeNode
}

func (r *kdTreeNode) Height() int {
	var treeHeight func(*kdTreeNode) int
	treeHeight = func(node *kdTreeNode) int {
		if node == nil {
			return 0
		}
		left := 1 + treeHeight(node.left)
		right := 1 + treeHeight(node.right)
		if left > right {
			return left
		} else {
			return right
		}
	}
	return treeHeight(r)
}

func (r *kdTreeNode) MinHeight() int {
	var treeHeight func(*kdTreeNode) int
	treeHeight = func(node *kdTreeNode) int {
		if node == nil {
			return 0
		}
		left := 1 + treeHeight(node.left)
		right := 1 + treeHeight(node.right)
		if left < right {
			return left
		} else {
			return right
		}
	}
	return treeHeight(r)
}

func (r *kdTreeNode) Prune(level int) {
	minHeight := r.MinHeight()
	if level >= minHeight {
		log.Println("Cannot prune the tree with more levels than minHeight ", minHeight)
		return
	}
	target := level - 1
	var dfs func(*kdTreeNode, int)
	dfs = func(node *kdTreeNode, level int) {
		if level == target {
			node.left = nil
			node.right = nil
		} else {
			dfs(node.left, level+1)
			dfs(node.right, level+1)
		}
	}
	dfs(r, 0)
}

func (r *kdTreeNode) GetLeafIndex(tuples []DatasetTuple) []int {
	var dfs func(*kdTreeNode, int, DatasetTuple) int
	dfs = func(node *kdTreeNode, id int, tup DatasetTuple) int {
		if node == nil { // leaf
			return id
		}
		if tup.Data[node.dim] <= node.value {
			return dfs(node.left, (id<<1)|0, tup)
		} else {
			return dfs(node.right, (id<<1)|1, tup)
		}
	}
	results := make([]int, 2<<uint(r.Height()-1))
	for _, tup := range tuples {
		results[dfs(r, 0, tup)] += 1
	}
	return results
}

func (r kdTreeNode) String() string {
	var myString func(*kdTreeNode, string) string
	myString = func(node *kdTreeNode, pad string) string {
		if node == nil {
			return ""
		} else {
			return fmt.Sprintf("%s(%d - %.5f)\n",
				pad, node.dim, node.value) +
				myString(node.left, pad+"\t") +
				myString(node.right, pad+"\t")
		}
	}
	return myString(&r, "")
}

// partitions the tuples and stores the tree structure in the kdTreeNode ptr
func NewKDTreePartition(tuples []DatasetTuple) *kdTreeNode {
	findMedian :=
		func(tuples []DatasetTuple, dim int) (float64, []DatasetTuple, []DatasetTuple) {
			// FIXME: this can be downgraded to O(n) time
			values := make([]float64, 0)
			for _, t := range tuples {
				values = append(values, t.Data[dim])
			}
			sort.Float64s(values)
			median := values[len(values)/2]
			left, right := make([]DatasetTuple, 0), make([]DatasetTuple, 0)
			for _, t := range tuples {
				if t.Data[dim] <= median {
					left = append(left, t)
				} else {
					right = append(right, t)
				}
			}
			return median, left, right
		}
	var partition func([]DatasetTuple, int, *kdTreeNode)
	partition = func(tuples []DatasetTuple, dim int, node *kdTreeNode) {
		tupSize := len(tuples[0].Data)
		median, left, right := findMedian(tuples, dim%tupSize)
		node.dim = dim % tupSize
		node.value = median
		if len(left) > 0 && len(right) > 0 {
			node.right = new(kdTreeNode)
			node.left = new(kdTreeNode)
			partition(left, dim+1, node.left)
			partition(right, dim+1, node.right)
		}
	}
	node := new(kdTreeNode)
	partition(tuples, 0, node)
	return node

}
