package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Class used to execute online indexing. The user can supply a map containing
// distances from original datasets and the indexer returns the coordinates of the
// specified dataset.
type OnlineIndexer struct {
	coordinates []DatasetCoordinates       // the coordinates of the datasets
	estimator   DatasetSimilarityEstimator // estimator object to calculate distances
	script      string                     // script to evaluate the acceptable solutions

	dimensionality int // the number of dimensions of the coordinates
}

// NewOnlineIndexer is a constructor function used to initialize an OnlineIndexer
// object.
func NewOnlineIndexer(estimator DatasetSimilarityEstimator,
	coordinates []DatasetCoordinates,
	script string) *OnlineIndexer {
	indexer := new(OnlineIndexer)
	indexer.coordinates = coordinates
	indexer.estimator = estimator
	indexer.script = script

	indexer.dimensionality = len(coordinates[0])
	return indexer
}

// Calculate method is responsible to calculate the coordinates of the specified
// dataset. In case that such a dataset cannot be represented by the specified
// coordinates system, an error is returned.
func (o *OnlineIndexer) Calculate(dataset *Dataset) (DatasetCoordinates, error) {
	// calculate the distances for the new dataset
	log.Println("Picking random datasets")
	perm := rand.Perm(len(o.estimator.Datasets()))

	writer, err := ioutil.TempFile("/tmp", "coordinates.csv-")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer writer.Close()
	defer os.Remove(writer.Name())

	for i := 0; i < o.dimensionality; i++ {
		fmt.Fprintf(writer, "x%d,", i+1)
	}
	fmt.Fprintf(writer, "d\n")

	for i := 0; i < o.dimensionality; i++ {
		idx := perm[i]
		dat := o.estimator.Datasets()[idx]
		sim := o.estimator.Similarity(dataset, dat)

		for j := 0; j < o.dimensionality; j++ {
			fmt.Fprintf(writer, "%.5f,", o.coordinates[idx][j])
		}
		fmt.Fprintf(writer, "%.5f\n", 1.0/sim)
	}
	log.Println(o.solveQuadSystem(writer.Name()))
	return nil, nil
}

// solveQuadSystem solves the quadratic polynomial system in order to identify
// the coordinates of the new point.
func (o *OnlineIndexer) solveQuadSystem(fileName string) ([]DatasetCoordinates, error) {
	log.Println("Executing", o.script, "with argument", fileName)
	// execute script
	cmd := exec.Command(o.script, fileName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	results := make([]DatasetCoordinates, 0)
	buffer := bytes.NewBuffer(output)
	l, _ := buffer.ReadString('\n')
	solutions, err := strconv.Atoi(strings.TrimSpace(l))
	if err != nil {
		return nil, err
	}
	for i := 0; i < solutions; i++ {
		l, err = buffer.ReadString('\n')
		if err != nil {
			return nil, err
		}

		ar := strings.Split(l, " ")
		solution := make(DatasetCoordinates, o.dimensionality)
		for j := range solution {
			solution[j], err = strconv.ParseFloat(strings.TrimSpace(ar[j]), 64)
			if err != nil {
				return nil, err
			}

		}
		results = append(results, solution)
	}
	return results, nil
}
