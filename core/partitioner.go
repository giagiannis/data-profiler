package core

import (
	"bytes"
	"errors"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// DataPartitioner is responsible to partition a dataset and, upon estimating,
// the basic partitioning scheme, dynamically partition new datasets.
type DataPartitioner interface {

	// Construct estimates the partitioning of the provided tuples (offline)
	Construct([]DatasetTuple) error
	// Partition executes partitioning to new datasets
	Partition([]DatasetTuple) ([][]DatasetTuple, error)
	// Configure provides the necessary configuration option to DataPartitioner
	Configure(map[string]string)
	// Options returns a list of options the DataPartitioner accepts with a
	// description
	Options() map[string]string
	// Serialize converts a DataPartitioner object to a stream of bytes
	Serialize() []byte
	// Deserialize converts a stream of bytes to a DataPartitioner object
	Deserialize([]byte)
}

// DataPartitionerType represents the type of the DataPartitioner struct
type DataPartitionerType uint8

const (
	// DataPartitionerKDTree utilizes a kd-tree for partitioning
	DataPartitionerKDTree DataPartitionerType = iota + 1
	// DataPartitionerKMeans utilizes kmeans for partitioning
	DataPartitionerKMeans DataPartitionerType = iota + 2
)

const KMeansMaxIteration = 10000

// NewDataPartitioner is the factory method for the creation of a new
// DataPartitioner object
func NewDataPartitioner(dpType DataPartitionerType, conf map[string]string) DataPartitioner {
	var obj DataPartitioner
	if dpType == DataPartitionerKDTree {
		obj = new(KDTreePartitioner)
	} else if dpType == DataPartitionerKMeans {
		obj = new(KMeansPartitioner)
	} else {
		return nil
	}
	obj.Configure(conf)
	return obj
}

// DeserializePartitioner returns instantiates a new partitioner from a serialized
// version
func DeserializePartitioner(b []byte) DataPartitioner {
	buff := bytes.NewBuffer(b)
	tempInt := make([]byte, 4)
	buff.Read(tempInt)
	t := uint8(getIntBytes(tempInt))
	var res DataPartitioner
	if t == uint8(DataPartitionerKMeans) {
		res = new(KMeansPartitioner)
	} else if t == uint8(DataPartitionerKDTree) {
		res = new(KDTreePartitioner)
	}
	res.Deserialize(b)
	return res
}

// KMeansPartitioner applies the k-means clustering algorithm to a given dataset
// and using the calculated centroids, it partitions newly provided datasets
// according to their distance from them
type KMeansPartitioner struct {
	// k of k-means
	k int
	// the centroids of the clusters
	centroids []DatasetTuple
	// the weights of the columns - used for distance normalization
	weights []float64
}

// Options returns the configuration options of the KMeansPartitioner
func (p *KMeansPartitioner) Options() map[string]string {
	return map[string]string{
		"partitions": "the number of partitions to use (k)",
		"weights": "the weights of the columns to utilize for the comparison" +
			"(default is to 1/(max - min) for each column)",
	}
}

// Configure provides the necessary configuration options to the
// KMeansPartitioner struct
func (p *KMeansPartitioner) Configure(conf map[string]string) {
	if val, ok := conf["partitions"]; ok {
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			p.k = int(v)
		}
	} else {
		log.Println("Setting default k value")
		p.k = 1
	}

	if val, ok := conf["weights"]; ok {
		arr := strings.Split(val, ",")
		p.weights = make([]float64, len(arr))
		for i := range arr {
			v, err := strconv.ParseFloat(arr[i], 64)
			if err != nil {
				log.Println(err)
			} else {
				p.weights[i] = v
			}
		}
	}
}

// initializeCentroids estimates a very first position of the centroids
// FIXME: can consider kmeans++
func (p *KMeansPartitioner) initializeCentroids(tuples []DatasetTuple) {
	perm := rand.Perm(len(tuples))
	p.centroids = make([]DatasetTuple, p.k)
	for i := 0; i < p.k; i++ {
		p.centroids[i] = tuples[perm[i]]
	}
}

// estimateWeights estimates the weights to be utilized in measuring the distances
func (p *KMeansPartitioner) estimateWeights(tuples []DatasetTuple) {
	if len(tuples) == 0 {
		return
	}
	maxValues := make([]float64, len(tuples[0].Data))
	minValues := make([]float64, len(tuples[0].Data))
	for i := range maxValues {
		maxValues[i] = math.Inf(-1)
		minValues[i] = math.Inf(1)
	}

	for _, t := range tuples {
		for j, v := range t.Data {
			if v >= maxValues[j] {
				maxValues[j] = v
			}
			if v <= minValues[j] {
				minValues[j] = v
			}
		}
	}
	p.weights = make([]float64, len(tuples[0].Data))
	for i := range maxValues {
		if maxValues[i] == minValues[i] {
			p.weights[i] = 0.0
		} else {
			p.weights[i] = 1.0 / ((maxValues[i] - minValues[i]) * (maxValues[i] - minValues[i]))
		}
	}
}

// assignTuplesToCentroids partitions the tuples according to their
// weighted distance from the calculated centroids
func (p *KMeansPartitioner) assignTuplesToCentroids(tuples []DatasetTuple) [][]DatasetTuple {
	groups := make([][]DatasetTuple, len(p.centroids))
	for _, t := range tuples {
		closestIdx, closestDst := 0, p.distance(t, p.centroids[0])
		for j, c := range p.centroids {
			currentDst := p.distance(t, c)
			if p.distance(t, c) < closestDst {
				closestIdx, closestDst = j, currentDst
			}
		}
		groups[closestIdx] = append(groups[closestIdx], t)
	}
	return groups
}

// estimateCentroids estimates the new centroids based on the given clusters
func (p *KMeansPartitioner) estimateCentroids(clusters [][]DatasetTuple) []DatasetTuple {
	var centroids []DatasetTuple
	for _, c := range clusters {
		var centroid []float64
		if len(clusters) > 0 {
			centroid = make([]float64, len(c[0].Data))
		}
		for _, t := range c {
			for i, v := range t.Data {
				centroid[i] += v
			}
		}
		for i := range centroid {
			centroid[i] = centroid[i] / float64(len(c))
		}
		centroids = append(centroids, DatasetTuple{centroid})
	}
	return centroids
}

// distance returns the weighted distance between two tuples
func (p *KMeansPartitioner) distance(a, b DatasetTuple) float64 {
	sum := 0.0
	for i := range a.Data {
		diff := (a.Data[i] - b.Data[i])
		sum += p.weights[i] * diff * diff
	}
	return math.Sqrt(sum)
}

// centroidsDelta returns the difference between the new and old centroids
func (p *KMeansPartitioner) centroidsDelta(a, b []DatasetTuple) float64 {
	if len(a) != len(b) {
		return math.NaN()
	}
	sum := 0.0
	for i := range a {
		sum += p.distance(a[i], b[i])
	}
	return sum
}

// Construct runs the k-means algorithm and estimates the centroids of the
// cluster (in order to be later used for partitioning.
func (p *KMeansPartitioner) Construct(tuples []DatasetTuple) error {
	if tuples == nil || len(tuples) == 0 {
		return errors.New("No tuples provided")
	}
	if p.weights == nil { // need to set weights
		p.estimateWeights(tuples)
	}

	p.initializeCentroids(tuples)
	delta := math.Inf(1)
	for i := 0; i < KMeansMaxIteration && delta > 0; i++ {
		clusters := p.assignTuplesToCentroids(tuples)
		newCentroids := p.estimateCentroids(clusters)
		delta = p.centroidsDelta(p.centroids, newCentroids)
		p.centroids = newCentroids
	}
	return nil
}

// Partition receives a set of tuples as input and returns a number of clusters
func (p *KMeansPartitioner) Partition(tuples []DatasetTuple) (
	[][]DatasetTuple, error) {
	if len(tuples) == 0 {
		return nil, errors.New("no tuples to partition")
	}
	if p.centroids == nil || len(p.centroids) == 0 {
		return nil, errors.New("centroids not estimated")
	}
	if len(tuples[0].Data) != len(p.centroids[0].Data) {
		return nil, errors.New("wrong data dimensionality")
	}
	return p.assignTuplesToCentroids(tuples), nil
}

func (p *KMeansPartitioner) Serialize() []byte {
	buffer := new(bytes.Buffer)
	buffer.Write(getBytesInt(int(DataPartitionerKMeans)))
	buffer.Write(getBytesInt(p.k))
	buffer.Write(getBytesInt(len(p.weights)))
	for i := range p.weights {
		buffer.Write(getBytesFloat(p.weights[i]))
	}
	for i := range p.centroids {
		for j := range p.centroids[i].Data {
			buffer.Write(getBytesFloat(p.centroids[i].Data[j]))
		}
	}
	return buffer.Bytes()
}

func (p *KMeansPartitioner) Deserialize(b []byte) {
	buff := bytes.NewBuffer(b)
	bytesInt := make([]byte, 4)
	bytesFloat := make([]byte, 8)
	buff.Read(bytesInt)	// consume kmeans
	buff.Read(bytesInt)
	p.k = getIntBytes(bytesInt)
	buff.Read(bytesInt)
	tupleDimensionality := getIntBytes(bytesInt)
	p.weights = make([]float64, tupleDimensionality)
	for i := range p.weights {
		buff.Read(bytesFloat)
		p.weights[i] = getFloatBytes(bytesFloat)
	}
	p.centroids = make([]DatasetTuple, p.k)
	for i := 0; i < p.k; i++ {
		p.centroids[i] = *new(DatasetTuple)
		p.centroids[i].Data = make([]float64, tupleDimensionality)
		for j := 0; j < tupleDimensionality; j++ {
			buff.Read(bytesFloat)
			p.centroids[i].Data[j] = getFloatBytes(bytesFloat)
		}
	}
}

// KDTreePartitioner generates a kd-tree on the selected columns
// and applies the partitioning to new datasets
type KDTreePartitioner struct {
	// the number of partitions to create
	partitions int
	// the datset columns to consider for the partitioning
	columns []int
	// kdtree is the tree structure
	kdtree []*treeNode
}

// Options returns a list of options
func (p *KDTreePartitioner) Options() map[string]string {
	return map[string]string{
		"partitions": "the number of partitions",
		"columns":    "comma separated list of column ids to use (default is all)",
	}
}

// Configure provides the necessary configuration params
func (p *KDTreePartitioner) Configure(conf map[string]string) {
	if val, ok := conf["partitions"]; ok {
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			p.partitions = int(v)
		}
	} else {
		log.Println("Setting default k value")
		p.partitions = 1
	}

	if val, ok := conf["columns"]; ok {
		arr := strings.Split(val, ",")
		p.columns = make([]int, len(arr))
		for i := range arr {
			v, err := strconv.ParseInt(arr[i], 10, 32)
			if err != nil {
				log.Println(err)
			} else {
				p.columns[i] = int(v)
			}
		}
	}

}

// treeNode represents a node of the kd-tree
type treeNode struct {
	dim   int
	value float64
}

func (p *KDTreePartitioner) partition(column int, tuples []DatasetTuple) ([]DatasetTuple, []DatasetTuple, float64) {
	var left, right []DatasetTuple
	values := make([]float64, len(tuples))
	for i := range tuples {
		values[i] = tuples[i].Data[column]
	}
	goesLeft := true
	median := Percentile(values, 50)
	for i := range tuples {
		if tuples[i].Data[column] < median {
			left = append(left, tuples[i])
		} else if tuples[i].Data[column] > median {
			right = append(right, tuples[i])
		} else { // equality
			if goesLeft {
				left = append(left, tuples[i])
			} else {
				right = append(left, tuples[i])
			}
			goesLeft = !goesLeft
		}
	}
	return left, right, median
}

func (p *KDTreePartitioner) Construct(tuples []DatasetTuple) error {
	if tuples == nil || len(tuples) == 0 {
		return errors.New("tuples not provided")
	}
	if p.columns == nil {
		p.columns = make([]int, len(tuples[0].Data))
		for i := range p.columns {
			p.columns[i] = i
		}
	}
	p.kdtree = make([]*treeNode, p.partitions-1)
	var createKDTree func(int, int, []DatasetTuple)
	createKDTree = func(colIdx, treeIdx int, tuples []DatasetTuple) {
		if treeIdx >= len(p.kdtree) {
			return
		}
		if len(tuples) == 0 || tuples == nil {
			p.kdtree[treeIdx] = nil
			return
		}
		l, r, median := p.partition(p.columns[colIdx%len(p.columns)], tuples)
		p.kdtree[treeIdx] = &treeNode{dim: p.columns[colIdx%len(p.columns)], value: median}
		if len(l) > 0 && len(r) > 0 {
			createKDTree(colIdx+1, 2*treeIdx+1, l)
			createKDTree(colIdx+1, 2*treeIdx+2, r)
		} else {
			createKDTree(colIdx+1, 2*treeIdx+1, nil)
			createKDTree(colIdx+1, 2*treeIdx+2, nil)
		}
		return
	}
	createKDTree(0, 0, tuples)
	return nil
}

// Partition applies the previously constructed kd-tree in order to partition
// the given dataset
func (p *KDTreePartitioner) Partition(tuples []DatasetTuple) ([][]DatasetTuple, error) {
	if len(tuples) == 0 {
		return nil, errors.New("no tuples to partition")
	}
	if p.kdtree == nil {
		return nil, errors.New("kdtree not estimated")
	}

	maxHeight := int(math.Floor(math.Log(float64(p.partitions)) + 1))
	clusters := make([][]DatasetTuple, p.partitions)
	for _, t := range tuples {
		id, idx, bitOps := 0, 0, 0
		for idx < len(p.kdtree) {
			cur := p.kdtree[idx]
			bitOps += 1
			if t.Data[cur.dim] <= cur.value {
				id = (id << 1) | 0
				idx = 2*idx + 1
			} else {
				id = (id << 1) | 1
				idx = 2*idx + 2
			}
		}
		if bitOps < maxHeight {
			id = (id << 1) | 0
		}
		clusters[id] = append(clusters[id], t)
	}
	return clusters, nil
}

// Serialize returns a byte array with the serialized object
func (p *KDTreePartitioner) Serialize() []byte {
	buf := new(bytes.Buffer)
	buf.Write(getBytesInt(int(DataPartitionerKDTree)))
	buf.Write(getBytesInt(p.partitions))
	buf.Write(getBytesInt(len(p.columns)))
	for i := range p.columns {
		buf.Write(getBytesInt(p.columns[i]))
	}
	buf.Write(getBytesInt(len(p.kdtree)))
	for i := range p.kdtree {
		buf.Write(getBytesInt(p.kdtree[i].dim))
		buf.Write(getBytesFloat(p.kdtree[i].value))
	}
	return buf.Bytes()
}

// Deserialize parses a byte array and instantiates a new kdtree part. object
func (p *KDTreePartitioner) Deserialize(b []byte) {
	buff := bytes.NewBuffer(b)
	tempInt, tempFloat := make([]byte, 4), make([]byte, 8)
	buff.Read(tempInt)	// consume partitioner type
	buff.Read(tempInt)
	p.partitions = getIntBytes(tempInt)
	buff.Read(tempInt)
	cols := getIntBytes(tempInt)
	p.columns = make([]int, cols)
	for i := range p.columns {
		buff.Read(tempInt)
		p.columns[i] = getIntBytes(tempInt)
	}
	buff.Read(tempInt)
	count := getIntBytes(tempInt)
	p.kdtree = make([]*treeNode, count)
	for i := range p.kdtree {
		buff.Read(tempInt)
		buff.Read(tempFloat)
		p.kdtree[i] = &treeNode{dim: getIntBytes(tempInt), value: getFloatBytes(tempFloat)}
	}
}
