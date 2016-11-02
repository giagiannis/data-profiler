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

type PartitionerType uint8

const (
	UNIFORM PartitionerType = iota + 1
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

	if a.partitionType == UNIFORM {
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
func DiscoverDatasets(inputDir string) []Dataset {
	log.Println("Discovering datasets")
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return nil
	}
	datasets := make([]Dataset, len(files))
	for i, f := range files {
		datasets[i] = *NewDataset(inputDir + "/" + f.Name())
	}
	return datasets

}
