package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// DatasetCoordinates is a struct for representing the dataset coordinates
type DatasetCoordinates []float64

// MDScaling is responsible for the execution of a MultiDimensional Scaling
// algorithm in order to provide coefficients for each dataset, based on a
// a similarity matrix.
type MDScaling struct {
	script string               // script to be used for the execution
	k      int                  // number of output dimensions
	matrix *DatasetSimilarities // the similarity matrix

	coordinates []DatasetCoordinates // the coordinates matrix
	stress      float64              // the stress factor
}

// NewMDScaling is the default MDScaling constructor; it initializes a new
// MDScaling object, based on the provided DatasetSimilarities struct and the
// k factor that determines the number of target dimensions. If k<1, then
// auto estimation takes place.
func NewMDScaling(matrix *DatasetSimilarities, k int, script string) *MDScaling {
	mds := new(MDScaling)
	mds.matrix = matrix
	mds.k = k
	mds.script = script
	return mds
}

// Compute functions executes the Multidimensional Scaling computation.
func (md *MDScaling) Compute() error {
	// create a temp file containing similarity matrix as a csv
	datasets := md.matrix.Datasets()
	writer, err := ioutil.TempFile("/tmp", "similarities")
	if err != nil {
		return err
	}
	defer writer.Close()
	//defer os.Remove(writer.Name())
	for _, d1 := range datasets {
		for j, d2 := range datasets {
			val := md.matrix.Get(d1.Path(), d2.Path())
			writer.WriteString(fmt.Sprintf("%.5f", val))
			if j < len(datasets)-1 {
				writer.WriteString(",")
			}
		}
		writer.WriteString("\n")
	}

	// execute computation
	if md.k < 1 || md.k > len(md.matrix.Datasets())-1 { // binary search in the interval [1, n-1]
		return errors.New("K factor must be between [1, n-1], n being the # of datasets")
	} else { // execute solution
		md.coordinates, md.stress, err = md.executeScript(writer.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

// Coordinates getter returns the dataset coordinates in a nxk slice (n being
// the number of datasets).
func (md *MDScaling) Coordinates() []DatasetCoordinates {
	return md.coordinates
}

// Stress getter returns the stress factor for the calculated solution
func (md *MDScaling) Stress() float64 {
	return md.stress
}

// Variances function returns the variances of the principal coordinates
func (md *MDScaling) Variances() ([]float64, error) {
	if md.coordinates == nil {
		return nil, errors.New("Coordinates not calculated")
	}
	variances := make([]float64, md.k)
	for i := 0; i < md.k; i++ {
		max, min := md.coordinates[0][i], md.coordinates[0][i]
		for j := 0; j < len(md.coordinates); j++ {
			if md.coordinates[j][i] > max {
				max = md.coordinates[j][i]
			}
			if md.coordinates[j][i] < min {
				min = md.coordinates[j][i]
			}
		}
		variances[i] = math.Abs(max - min)
	}
	return variances, nil
}

// executeScript is used to execute the specified MDS script and parse
// its results. Not part of the structs public API. If successful, returns
// the coordinates slice, the stress factor and nil errors; If not successful,
// returns nil results and an error object
//
func (md *MDScaling) executeScript(smPath string) ([]DatasetCoordinates, float64, error) {
	command := exec.Command(md.script, smPath, strconv.Itoa(md.k))
	buf, err := command.CombinedOutput()
	if err != nil {
		log.Println(string(buf))
		return nil, 0.0, err
	}
	lines := strings.Split(string(buf), "\n")
	// parse stress float
	stress, err := strconv.ParseFloat(lines[0], 64)
	if err != nil {
		return nil, 0.0, err
	}
	count := 0
	for i := 1; i < len(lines) && len(lines[i]) > 0; i++ {
		if len(lines[i]) > 0 {
			count += 1
		}
	}
	// parse coordinates slices
	coordinates := make([]DatasetCoordinates, count)
	for i := 1; i < len(lines) && len(lines[i]) > 0; i++ {
		splitLine := strings.Split(lines[i], " ")
		coordinates[i-1] = make([]float64, md.k)
		for j := 0; j < md.k; j++ {
			coordinates[i-1][j], err = strconv.ParseFloat(splitLine[j], 64)
			if err != nil {
				return nil, 0.0, err
			}
		}
	}
	return coordinates, stress, nil
}
