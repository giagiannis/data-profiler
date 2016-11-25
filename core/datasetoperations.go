package core

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

// DatasetsIntersection function is used to calculate the intersection
// of two datasets and returns the tuples that belong to it.
func DatasetsIntersection(a, b *Dataset) []DatasetTuple {
	a.ReadFromFile()
	b.ReadFromFile()
	dict := make(map[string]bool)
	for _, dt := range a.Data() {
		dict[dt.Serialize()] = true
	}
	result := make([]DatasetTuple, 0)
	for _, dt := range b.Data() {
		ok := dict[dt.Serialize()]
		if ok {
			result = append(result, dt)
		}
	}

	return result
}

// DatasetsUnion function is used to calculate the union of two datasets
// and returns the tuples that belong to it.
func DatasetsUnion(a, b *Dataset) []DatasetTuple {
	a.ReadFromFile()
	b.ReadFromFile()
	dict := make(map[string]bool)
	for _, dt := range a.Data() {
		dict[dt.Serialize()] = true
	}
	for _, dt := range b.Data() {
		dict[dt.Serialize()] = true
	}
	result := make([]DatasetTuple, 0)
	for k, _ := range dict {
		t := new(DatasetTuple)
		t.Deserialize(k)
		result = append(result, *t)
	}

	return result
}

type PartitionerType uint8

const (
	PARTITIONER_TYPE_UNIFORM PartitionerType = iota + 1
)

// DatasetPartitioner accepts a single dataset and it is responsible to
// partition it.
type DatasetPartitioner struct {
	input         string          // the input file to partition
	output        string          // the output dir that holds the partitions
	splits        int             // number of files to generate
	partitionType PartitionerType // type of the partitioner
}

func NewDatasetPartitioner(input, output string, splits int, partitionType PartitionerType) *DatasetPartitioner {
	a := new(DatasetPartitioner)
	a.input = input
	a.output = output
	a.splits = splits
	a.partitionType = partitionType
	return a
}

// Delete function deletes the output directory, containing the Dataset splits
func (a *DatasetPartitioner) Delete() {
	os.RemoveAll(a.output)
}

// Partition function is used to execute the partitioning
func (a *DatasetPartitioner) Partition() {
	os.Mkdir(a.output, 0777)
	newFiles := make([]*os.File, a.splits)
	for i := 0; i < a.splits; i++ {
		fileName := fmt.Sprintf("%s/split-%d", a.output, i)
		newFiles[i], _ = os.Create(fileName)
	}
	file, _ := os.Open(a.input)
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	header := scanner.Text()
	for _, f := range newFiles {
		f.WriteString(header + "\n")
	}

	if a.partitionType == PARTITIONER_TYPE_UNIFORM {
		a.uniform(scanner, newFiles)
	}

	for _, f := range newFiles {
		f.Close()
	}
	file.Close()

}

func (a *DatasetPartitioner) uniform(scanner *bufio.Scanner, newFiles []*os.File) {
	rand.Seed(int64(time.Now().Nanosecond()))
	for scanner.Scan() {
		fileChosen := rand.Int() % a.splits
		newFiles[fileChosen].WriteString(scanner.Text() + "\n")
	}
}

// DiscoverDatasets is used to return a slice of Datasets when a new
// splits directory is provided
func DiscoverDatasets(inputDir string) []*Dataset {
	log.Println("Discovering datasets")
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return nil
	}
	datasets := make([]*Dataset, len(files))
	for i, f := range files {
		datasets[i] = NewDataset(inputDir + "/" + f.Name())
	}
	return datasets

}
