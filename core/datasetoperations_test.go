package core

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestDatatesetIntersection(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 2, 10)
	res := DatasetsIntersection(datasets[0], datasets[1])
	if len(datasets[0].Data()) < len(res) ||
		len(datasets[1].Data()) < len(res) {
		t.Log(len(datasets[0].Data()))
		t.Log(len(datasets[1].Data()))
		t.FailNow()
	}

	for _, d := range datasets {
		os.Remove(d.Path())
	}
}

func TestDatasetUnion(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 2, 10)
	res := DatasetsUnion(datasets[0], datasets[1])
	if len(datasets[0].Data()) > len(res) ||
		len(datasets[1].Data()) > len(res) {
		t.Log(len(datasets[0].Data()))
		t.Log(len(datasets[1].Data()))
		t.FailNow()
	}

	for _, d := range datasets {
		os.Remove(d.Path())
	}
}

func TestUniformPartition(t *testing.T) {
	//log.SetOutput(ioutil.Discard)
	numberOfDatasets := rand.Int()%200 + 1
	rand.Seed(int64(time.Now().Nanosecond()))
	part := NewDatasetPartitioner(TRAINSET, TRAINSET+"-splits/", numberOfDatasets, PARTITIONER_TYPE_UNIFORM)
	part.Partition()

	files := DiscoverDatasets(part.output)
	if len(files) != numberOfDatasets {
		t.Log("Different number of requested and created splits")
		t.FailNow()
	}

	lines := 1
	for _, f := range files {
		file, _ := os.Open(f.Path())
		lines += lineCounter(file)
		lines -= 1
		file.Close()
	}
	originalFile, _ := os.Open(part.input)
	original := lineCounter(originalFile)

	if lines != original {
		t.Log("Data points lost during the partition")
		t.FailNow()
	}

	part.Delete()
}

func lineCounter(f *os.File) int {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := f.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count

		case err != nil:
			return count
		}
	}
}
