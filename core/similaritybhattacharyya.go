package core

import (
	"errors"
	"fmt"
	"log"
	"sort"
)

type BhattacharyyaEstimator struct {
	datasets     []*Dataset           // datasets slice
	concurrency  int                  // the max number of threads that run in parallel
	similarities *DatasetSimilarities // the similarities struct
}

func (e *BhattacharyyaEstimator) Compute() error {
	// allocation of similarities struct
	e.similarities = NewDatasetSimilarities(e.datasets)
	log.Println("Fetching datasets in memory")
	if e.datasets == nil || len(e.datasets) == 0 {
		log.Println("No datasets were given")
		return errors.New("Empty dataset slice")
	}
	for _, d := range e.datasets {
		d.ReadFromFile()
	}

	root := new(kdTreeNode)
	kdTreePartition(e.datasets[0].Data(), 0, root)
	fmt.Println(root)
	fmt.Println(root.Height())
	// FIXME: the created partitions need to mean something here
	return nil
}

func (e *BhattacharyyaEstimator) GetSimilarities() *DatasetSimilarities {
	return e.similarities
}

type kdTreeNode struct {
	dim   int
	value float64
	right *kdTreeNode
	left  *kdTreeNode
}

func (r kdTreeNode) Height() int {
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
	return treeHeight(&r)
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
func kdTreePartition(tuples []DatasetTuple, dim int, node *kdTreeNode) {
	tupSize := len(tuples[0].Data)
	findMedian :=
		// FIXME: this can be downgraded to O(n) time
		func(tuples []DatasetTuple, dim int) (float64, []DatasetTuple, []DatasetTuple) {
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
	median, left, right := findMedian(tuples, dim%tupSize)
	node.dim = dim % tupSize
	node.value = median
	if len(left) > 1 && len(right) > 1 {
		node.right = new(kdTreeNode)
		node.left = new(kdTreeNode)
		kdTreePartition(left, dim+1, node.left)
		kdTreePartition(right, dim+1, node.right)
	}
}
