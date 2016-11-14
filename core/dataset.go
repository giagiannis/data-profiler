package core

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

const DATASET_SEPARATOR = ","

// Dataset struct represents a dataset object.
type Dataset struct {
	id     string
	path   string
	header []string
	data   []DatasetTuple
}

// NewDataset is the constructor for the Dataset struct. A random ID is assigned
// to a new dataset
func NewDataset(path string) *Dataset {

	d := new(Dataset)
	buffer := make([]byte, 4)
	rand.Read(buffer)
	d.id = fmt.Sprintf("%x", buffer)
	d.path = path
	return d
}

// Id getter for dataset
func (d Dataset) Id() string {
	return d.id
}

// Path getter for dataset
func (d Dataset) Path() string {
	return d.path
}

// Header getter for dataset - only works if ReadFromFile was successful
func (d Dataset) Header() []string {
	return d.header
}

// Data getter for dataset - only works if ReadFromFile was successful
func (d Dataset) Data() []DatasetTuple {
	return d.data
}

// String method for dataset object - returns the path of the dataset
func (d Dataset) String() string {
	return d.path
}

// Method used  to parse the Dataset into memory. If the data are previously read,
// the method is not re-executed.
func (d *Dataset) ReadFromFile() error {
	if d.Header() != nil && d.Data() != nil { // previously read
		return nil
	}
	dat, err := ioutil.ReadFile(d.path)
	if err != nil {
		return err
	}
	datSplit := strings.Split(fmt.Sprintf("%s", dat), "\n")
	if len(datSplit) < 1 {
		return errors.New("File without contents")
	}

	// reading header
	d.header = make([]string, 0)
	for _, s := range strings.Split(datSplit[0], DATASET_SEPARATOR) {
		d.header = append(d.header, s)
	}

	// reading data
	for i := 1; i < len(datSplit); i++ {
		if len(datSplit[i]) > 0 {
			t := new(DatasetTuple)
			t.Deserialize(datSplit[i])
			d.data = append(d.data, *t)
		}
	}

	return nil
}

// DatasetTuple represents a data tuple from the dataset
type DatasetTuple struct {
	Data []float64
}

// Deserializes a tuple from a string representation
func (t *DatasetTuple) Deserialize(data string) {
	for _, s := range strings.Split(data, DATASET_SEPARATOR) {
		v, _ := strconv.ParseFloat(s, 64)
		t.Data = append(t.Data, v)
	}
}

// Serializes the tuple to a string representation
func (t *DatasetTuple) Serialize() string {
	return t.String()
}

func (t DatasetTuple) String() string {
	max := len(t.Data) - 1
	buffer := new(bytes.Buffer)

	for i, v := range t.Data {
		buffer.WriteString(fmt.Sprintf("%.5f", v))
		if i < max {
			buffer.WriteString(", ")
		}
	}
	return fmt.Sprintf("%s", buffer.Bytes())
}
