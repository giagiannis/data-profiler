package core

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// BhattacharyyaEstimator is the similarity estimator that quantifies the similarity
// of the distribution between the datasets.
type BhattacharyyaEstimator struct {
	AbstractDatasetSimilarityEstimator

	// struct used to map dataset paths to indexes
	inverseIndex map[string]int
	// determines the height of the kd tree to be used
	maxPartitions int
	// hold the portion of the data examined for constructing the tree
	kdTreeSamplePerc float64
	// kd tree, utilized for dataset partitioning
	//	kdTree *kdTreeNode
	partitioner DataPartitioner
	// holds the number of points for each dataset region
	pointsPerRegion [][]int
	// holds the total number of points for each dataset
	datasetsSize []int
}

// Compute method constructs the Similarity Matrix
func (e *BhattacharyyaEstimator) Compute() error {
	return datasetSimilarityEstimatorCompute(e)
}

// Similarity returns the similarity between two datasets
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
		clusters, err := e.partitioner.Partition(a.Data())
		if err != nil {
			log.Println(err)
		}
		for _, c := range clusters {
			indexA = append(indexA, len(c))
		}
		//		indexA = e.kdTree.GetLeafIndex(a.Data())
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
		clusters, err := e.partitioner.Partition(b.Data())
		if err != nil {
			log.Println(err)
		}
		for _, c := range clusters {
			indexB = append(indexB, len(c))
		}

		//		indexB = e.kdTree.GetLeafIndex(b.Data())
		countB = len(b.Data())
	}
	return e.getValue(indexA, indexB, countA, countB)
}

// Configure sets a the configuration parameters of the estimator
func (e *BhattacharyyaEstimator) Configure(conf map[string]string) {
	if val, ok := conf["concurrency"]; ok {
		conv, err := strconv.ParseInt(val, 10, 32)
		e.concurrency = int(conv)
		if err != nil {
			log.Println(err)
		}
	} else {
		e.concurrency = 1
	}
	if val, ok := conf["partitions"]; ok {
		//conv, err := strconv.ParseInt(val, 10, 32)
		conv, err := strconv.ParseInt(val, 10, 32)
		e.maxPartitions = int(conv)
		if err != nil {
			log.Println(err)
		}
	} else {
		e.maxPartitions = 32
	}

	if val, ok := conf["dataset.sr"]; ok {
		//conv, err := strconv.ParseInt(val, 10, 32)
		conv, err := strconv.ParseFloat(val, 64)
		e.kdTreeSamplePerc = conv
		if err != nil {
			log.Println(err)
		}
	} else {
		e.kdTreeSamplePerc = 0.1
	}

	partitionerType := DataPartitionerKDTree
	if val, ok := conf["partitioner.type"]; ok {
		if "kmeans" == strings.ToLower(val) {
			partitionerType = DataPartitionerKMeans
		} else if "kdtree" == strings.ToLower(val) {
			partitionerType = DataPartitionerKDTree
		} else {
			log.Println("Unknown partitioner type, using default (kdtree)")
		}
	}

	partitionerConf := make(map[string]string)
	partitionerConf["partitions"] = fmt.Sprintf("%d", e.maxPartitions)
	// parse partitioner params
	log.Println(conf)
	for k, v := range conf {
		log.Println(k, v)
		if strings.HasPrefix(k, "partitioner.") {
			partitionerConf[strings.TrimPrefix(k, "partitioner.")] = v
		}
	}
	log.Println("Providing the following conf to the partitioner", partitionerConf)
	e.init(partitionerType, partitionerConf)

}

func (e *BhattacharyyaEstimator) init(partitionerType DataPartitionerType, partitionerConf map[string]string) {
	// initialization step
	e.inverseIndex = make(map[string]int)
	for i, d := range e.datasets {
		e.inverseIndex[d.Path()] = i
	}
	for _, d := range e.datasets {
		d.ReadFromFile()
	}
	//e.kdTree = newKDTreePartition(e.datasets[0].Data())
	s := e.sampledDataset()
	e.partitioner = NewDataPartitioner(partitionerType, partitionerConf)
	e.partitioner.Construct(s)
	e.pointsPerRegion = make([][]int, len(e.datasets))
	e.datasetsSize = make([]int, len(e.datasets))
	for i, d := range e.datasets {
		clusters, err := e.partitioner.Partition(d.Data())
		if err != nil {
			log.Println(err)
		} else {
			for _, c := range clusters {
				e.pointsPerRegion[i] = append(e.pointsPerRegion[i], len(c))
			}
			//			e.pointsPerRegion[i] = e.kdTree.GetLeafIndex(d.Data())
			e.datasetsSize[i] = len(d.Data())
		}
	}

	// UP TO THIS POINT
}

// Options returns a list of parameters that can be set by the user
func (e *BhattacharyyaEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency":      "max num of threads used (int)",
		"partitions":       "max number of partitions to be used for the estimation (default is 32)",
		"partitioner.type": "the partitioner type (one of kmeans, kdtree - default is kdtree) ",
		"partitioner.*":    "provide any argument to the partitioner instance using the partitioner.* prefix (e.g.: partitioner.weights=0.1,0.2 for kmeans)",
		"dataset.sr":       "determines the portion of datasets to sample for the partitioner construction",
		//	"columns":          "comma separated values of column indices to consider (starting from 0)  or all (default)",
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

// Serialize returns a byte array containing a serialized form of the estimator
func (e *BhattacharyyaEstimator) Serialize() []byte {
	buffer := new(bytes.Buffer)
	buffer.Write(getBytesInt(int(SimilarityTypeBhattacharyya)))

	bytes := datasetSimilarityEstimatorSerialize(e.AbstractDatasetSimilarityEstimator)
	buffer.Write(bytes)
	buffer.Write(getBytesInt(e.maxPartitions))

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
	serializedPartitioner := e.partitioner.Serialize()
	buffer.Write(getBytesInt(len(serializedPartitioner)))
	buffer.Write(serializedPartitioner)
	return buffer.Bytes()
}

// Deserialize constructs a similarity object based on the byte stream
func (e *BhattacharyyaEstimator) Deserialize(b []byte) {
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

	buffer.Read(tempInt)
	e.maxPartitions = getIntBytes(tempFloat)

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
	e.partitioner = DeserializePartitioner(tempCustom)
	//e.kdTree = new(kdTreeNode)
	//e.kdTree.Deserialize(tempCustom)

}

// sampledDataset returns a custom dataset that consist of the tuples of the
// previous
func (e *BhattacharyyaEstimator) sampledDataset() []DatasetTuple {
	log.Println("Generating a sampled and merged dataset with all tuples")
	var result []DatasetTuple
	for _, d := range e.datasets {
		tuplesToChoose := int(math.Floor(float64(len(d.Data())) * e.kdTreeSamplePerc))
		log.Printf("%d/%d tuples chosen for %s\n", tuplesToChoose, len(d.Data()), d.path)
		tuplesIdx := make(map[int]bool)
		for len(tuplesIdx) < tuplesToChoose {
			tuplesIdx[rand.Int()%len(d.Data())] = true
		}
		for k := range tuplesIdx {
			result = append(result, d.Data()[k])
		}
	}
	return result
}
