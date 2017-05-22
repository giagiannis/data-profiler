package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"time"
)

type BhattacharyyaEstimator struct {
	AbstractDatasetSimilarityEstimator

	// struct used to map dataset paths to indexes
	inverseIndex map[string]int
	// determines the height of the kd tree to be used
	kdTreeScaleFactor float64
	// kd tree, utilized for dataset partitioning
	kdTree *kdTreeNode
	// holds the number of points for each dataset region
	pointsPerRegion [][]int
	// holds the total number of points for each dataset
	datasetsSize []int
}

func (e *BhattacharyyaEstimator) Compute() error {
	// allocation of similarities struct
	e.similarities = NewDatasetSimilarities(len(e.datasets))

	log.Println("Fetching datasets in memory")
	if e.datasets == nil || len(e.datasets) == 0 {
		log.Println("No datasets were given")
		return errors.New("Empty dataset slice")
	}
	e.inverseIndex = make(map[string]int)
	for i, d := range e.datasets {
		log.Println(i, d.Path())
		e.inverseIndex[d.Path()] = i
		d.ReadFromFile()
	}

	start := time.Now()
	log.Println("Estimating a KD-tree partition")
	e.kdTree = NewKDTreePartition(e.datasets[0].Data())
	oldHeight := e.kdTree.Height()
	newHeight := int(float64(oldHeight) * e.kdTreeScaleFactor)
	log.Printf("Pruning the tree from height %d to %d\n", oldHeight, newHeight)
	e.kdTree.Prune(newHeight)
	e.pointsPerRegion = make([][]int, len(e.datasets))
	e.datasetsSize = make([]int, len(e.datasets))
	for i, d := range e.datasets {
		e.pointsPerRegion[i] = e.kdTree.GetLeafIndex(d.Data())
		e.datasetsSize[i] = len(d.Data())
	}

	datasetSimilarityEstimatorCompute(e)
	e.duration = time.Since(start).Seconds()
	return nil
}

func (e *BhattacharyyaEstimator) Similarity(a, b *Dataset) float64 {
	var indexA, indexB []int
	var countA, countB int
	if idx, ok := e.inverseIndex[a.Path()]; ok {
		indexA = e.pointsPerRegion[idx]
		countA = e.datasetsSize[idx]
	} else {
		err := a.ReadFromFile()
		if err != nil {
			log.Println(err)
		}
		indexA = e.kdTree.GetLeafIndex(a.Data())
		countA = len(a.Data())
	}

	if idx, ok := e.inverseIndex[b.Path()]; ok {
		indexB = e.pointsPerRegion[idx]
		countB = e.datasetsSize[idx]
	} else {
		err := b.ReadFromFile()
		if err != nil {
			log.Println(err)
		}
		indexB = e.kdTree.GetLeafIndex(b.Data())
		countB = len(b.Data())
	}
	return e.getValue(indexA, indexB, countA, countB)
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

func (e *BhattacharyyaEstimator) getValue(indA, indB []int, countA, countB int) float64 {
	sum := 0.0
	for k := 0; k < len(indA); k++ {
		sum += math.Sqrt(float64(indA[k] * indB[k]))
	}
	sum /= math.Sqrt(float64(countA * countB))
	return sum
}

func (e *BhattacharyyaEstimator) Serialize() []byte {
	buffer := new(bytes.Buffer)
	buffer.Write(getBytesInt(int(SIMILARITY_TYPE_BHATTACHARYYA)))

	bytes := datasetSimilarityEstimatorSerialize(e.AbstractDatasetSimilarityEstimator)
	buffer.Write(bytes)
	buffer.Write(getBytesFloat(e.kdTreeScaleFactor))

	// write points per region
	buffer.Write(getBytesInt(len(e.pointsPerRegion[0])))
	for i := range e.pointsPerRegion {
		for j := range e.pointsPerRegion[i] {
			buffer.Write(getBytesInt(e.pointsPerRegion[i][j]))
		}
	}
	// write datasets size
	for _, s := range e.datasetsSize {
		buffer.Write(getBytesInt(s))
	}
	// write kdtree
	serializedTree := e.kdTree.Serialize()
	buffer.Write(getBytesInt(len(serializedTree)))
	buffer.Write(e.kdTree.Serialize())
	return buffer.Bytes()
}

func (e *BhattacharyyaEstimator) Deserialize(b []byte) {
	// FIXME: implement the Deserialized method
	buffer := bytes.NewBuffer(b)
	tempInt := make([]byte, 4)
	tempFloat := make([]byte, 8)
	buffer.Read(tempInt) // contains estimator type

	var count int
	buffer.Read(tempInt)
	absEstBytes := make([]byte, getIntBytes(tempInt))
	buffer.Read(absEstBytes)
	e.AbstractDatasetSimilarityEstimator =
		*datasetSimilarityEstimatorDeserialize(absEstBytes)

	buffer.Read(tempFloat)
	e.kdTreeScaleFactor = getFloatBytes(tempFloat)

	e.inverseIndex = make(map[string]int)
	for i := range e.datasets {
		e.inverseIndex[e.datasets[i].Path()] = i
	}

	buffer.Read(tempInt)
	count = getIntBytes(tempInt)
	e.pointsPerRegion = make([][]int, len(e.datasets))
	for i := range e.pointsPerRegion {
		e.pointsPerRegion[i] = make([]int, count)
		for j := range e.pointsPerRegion[i] {
			buffer.Read(tempInt)
			e.pointsPerRegion[i][j] = getIntBytes(tempInt)
		}
	}

	e.datasetsSize = make([]int, len(e.datasets))
	for i := range e.datasetsSize {
		buffer.Read(tempInt)
		e.datasetsSize[i] = getIntBytes(tempInt)
	}

	buffer.Read(tempInt)
	count = getIntBytes(tempInt)
	tempCustom := make([]byte, count)
	buffer.Read(tempCustom)
	e.kdTree = new(kdTreeNode)
	e.kdTree.Deserialize(tempCustom)

}

type kdTreeNode struct {
	dim   int
	value float64
	right *kdTreeNode
	left  *kdTreeNode
}

func (r *kdTreeNode) Serialize() []byte {
	// FIXME: Serialize and Deserialize do not work for partial trees
	var countNodes func(*kdTreeNode) int
	countNodes = func(node *kdTreeNode) int {
		if node == nil {
			return 0
		}
		count := 0
		count += countNodes(node.left)
		count += countNodes(node.right)
		return count + 1
	}
	dataSize := 8 + 4
	nodes := countNodes(r)
	buf := make([]byte, dataSize*nodes)
	var serializeNode func(*kdTreeNode, int)
	serializeNode = func(node *kdTreeNode, offset int) {
		if node == nil {
			return
		}
		for i, b := range getBytesInt(node.dim) {
			buf[offset*dataSize+i] = b
		}
		count := len(getBytesInt(node.dim))
		for i, b := range getBytesFloat(node.value) {
			buf[offset*dataSize+count+i] = b
		}
		serializeNode(node.left, (2*offset + 1))
		serializeNode(node.right, (2*offset + 2))
	}
	serializeNode(r, 0)
	return buf
}

func (r *kdTreeNode) Deserialize(b []byte) {
	// FIXME: Serialize and Deserialize do not work for partial trees
	dataSize := 8 + 4
	var deserialize func(int) *kdTreeNode
	deserialize = func(offset int) *kdTreeNode {
		if offset*dataSize >= len(b) {
			return nil
		}
		node := new(kdTreeNode)
		tempInt := make([]byte, 4)
		for i := range tempInt {
			tempInt[i] = b[offset*dataSize+i]
		}
		node.dim = getIntBytes(tempInt)
		tempFloat := make([]byte, 8)
		for i := range tempFloat {
			tempFloat[i] = b[offset*dataSize+4+i]
		}
		node.value = getFloatBytes(tempFloat)
		node.left = deserialize(2*offset + 1)
		node.right = deserialize(2*offset + 2)
		return node
	}
	n := deserialize(0)
	r.dim = n.dim
	r.value = n.value
	r.left = n.left
	r.right = n.right
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
	//minHeight := r.MinHeight()
	//if level > minHeight {
	//	log.Println("Cannot prune the tree with more levels than minHeight ", minHeight)
	//	return
	//}
	target := level - 1
	var dfs func(*kdTreeNode, int)
	dfs = func(node *kdTreeNode, level int) {
		if node == nil {
			return
		}
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
	treeHeight := r.Height()
	var dfs func(*kdTreeNode, int, int, DatasetTuple) int
	dfs = func(node *kdTreeNode, id, level int, tup DatasetTuple) int {
		if node == nil { // leaf
			return id << uint(treeHeight-level)
		}
		if tup.Data[node.dim] <= node.value {
			return dfs(node.left, (id<<1)|0, level+1, tup)
		} else {
			return dfs(node.right, (id<<1)|1, level+1, tup)
		}
	}
	results := make([]int, 2<<uint(treeHeight-1))
	for _, tup := range tuples {
		results[dfs(r, 0, 0, tup)] += 1
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
