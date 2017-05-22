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
	script string                   // script to be used for the execution
	k      int                      // number of output dimensions
	matrix *DatasetSimilarityMatrix // the similarity matrix

	coordinates []DatasetCoordinates // the coordinates matrix
	gof         float64              // the gof factor
}

// NewMDScaling is the default MDScaling constructor; it initializes a new
// MDScaling object, based on the provided DatasetSimilarities struct and the
// k factor that determines the number of target dimensions. If k<1, then
// auto estimation takes place.
func NewMDScaling(matrix *DatasetSimilarityMatrix, k int, script string) *MDScaling {
	mds := new(MDScaling)
	mds.matrix = matrix
	mds.k = k
	mds.script = script
	return mds
}

// Compute functions executes the Multidimensional Scaling computation.
func (md *MDScaling) Compute() error {
	// create a temp file containing similarity matrix as a csv
	writer, err := ioutil.TempFile("/tmp", "similarities")
	if err != nil {
		return err
	}
	defer writer.Close()
	log.Println("Writing distances to file", writer.Name())
	//defer os.Remove(writer.Name())
	maxDistance := 0.0
	for i := 0; i < md.matrix.Capacity(); i++ {
		for j := i + 1; j < md.matrix.Capacity(); j++ {
			dist := SimilarityToDistance(md.matrix.Get(i, j))
			if !math.IsInf(dist, 0) && dist > maxDistance {
				maxDistance = dist
			}
		}
	}
	for i := 0; i < md.matrix.Capacity(); i++ {
		for j := 0; j < md.matrix.Capacity(); j++ {
			dist := SimilarityToDistance(md.matrix.Get(i, j))
			if math.IsInf(dist, 0) {
				dist = maxDistance
			}
			writer.WriteString(fmt.Sprintf("%.5f", dist))
			if j < md.matrix.Capacity()-1 {
				writer.WriteString(",")
			}
		}
		writer.WriteString("\n")
	}

	// execute computation
	if md.k < 1 || md.k > md.matrix.Capacity()-1 { // binary search in the interval [1, n-1]
		return errors.New("K factor must be between [1, n-1], n being the # of datasets")
	}
	// execute solution
	md.coordinates, md.gof, err = md.executeScript(writer.Name())
	if err != nil {
		return err
	}
	return nil
}

// Coordinates getter returns the dataset coordinates in a nxk slice (n being
// the number of datasets).
func (md *MDScaling) Coordinates() []DatasetCoordinates {
	return md.coordinates
}

// Gof getter returns the gof factor for the calculated solution
func (md *MDScaling) Gof() float64 {
	return md.gof
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
// the coordinates slice, the gof factor and nil errors; If not successful,
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
	// parse gof float
	gof, err := strconv.ParseFloat(lines[0], 64)
	if err != nil {
		return nil, 0.0, err
	}
	count := 0
	for i := 1; i < len(lines) && len(lines[i]) > 0; i++ {
		if len(lines[i]) > 0 {
			count++
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
	return coordinates, gof, nil
}
