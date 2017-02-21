package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/giagiannis/data-profiler/core"
)

type expOnlineIndexerParams struct {
	indexer  *core.OnlineIndexer
	datasets []*core.Dataset
	sm       *core.DatasetSimilarities
	coords   []core.DatasetCoordinates

	outputStress *string
	outputCoords *string
	compDatasets *int
}

func expOnlineIndexerParseParams() *expOnlineIndexerParams {
	params := new(expOnlineIndexerParams)
	coordinatesPath :=
		flag.String("c", "", "file that keeps the original coordinates")
	estimatorPath :=
		flag.String("e", "", "path of the serialized estimator object")
	saScript :=
		flag.String("s", "", "Script that executes Simulated Annealing")
	logger :=
		flag.String("l", "", "log file")
	params.outputStress =
		flag.String("os", "", "output for the stress values")
	params.outputCoords =
		flag.String("oc", "", "output for the coords file")
	inputDatasets :=
		flag.String("i", "", "datasets to index")
	params.compDatasets =
		flag.Int("com", 0, "datasets to be used for the comparison")

	flag.Parse()
	setLogger(*logger)
	if *coordinatesPath == "" || *estimatorPath == "" || *saScript == "" {
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// parse estimator
	f, err := os.Open(*estimatorPath)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	est := core.DeserializeSimilarityEstimator(buf)
	params.sm = est.GetSimilarities()

	// parse coordinates files
	f, err = os.Open(*coordinatesPath)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err = ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	lines := strings.Split(strings.TrimSpace(string(buf)), "\n")
	if len(lines) < 1 {
		log.Fatalln("Empty coordinates file")
	}
	var datasetCoordinates []core.DatasetCoordinates

	header := strings.Split(strings.TrimSpace(lines[0]), " ")
	dimensionality := len(header)
	for i := 1; i < len(lines); i++ {
		coords := make([]float64, dimensionality)
		lineArray := strings.Split(lines[i], " ")
		for j := range coords {
			val, err := strconv.ParseFloat(lineArray[j], 64)
			if err != nil {
				log.Println(err)
			} else {
				coords[j] = val
			}
		}
		datasetCoordinates = append(datasetCoordinates, coords)
	}
	params.coords = datasetCoordinates

	// discovering datasets
	if *inputDatasets != "" {
		params.datasets = core.DiscoverDatasets(*inputDatasets)
	}

	// construct indexer object
	if len(params.datasets) > 0 {
		params.indexer = core.NewOnlineIndexer(est, datasetCoordinates, *saScript)
		params.indexer.DatasetsToCompare(*params.compDatasets)
	}
	return params
}

func expOnlineIndexerRun() {
	params := expOnlineIndexerParseParams()
	initialStress := getTotalStress(params.sm, params.coords)
	log.Println("Initial total stress:", initialStress)
	if len(params.datasets) > 0 {
		for _, d := range params.datasets {
			coord, stress, err := params.indexer.Calculate(d)
			if err == nil {
				params.coords = append(params.coords, coord)
				initialStress += stress
				log.Println("Results:", coord, stress)
				log.Println("Current total stress:", initialStress)
			} else {
				log.Println(err)
			}
		}
		f := setOutput(*params.outputCoords)
		for i := 0; i < len(params.coords[0]); i++ {
			fmt.Fprintf(f, "x_%d ", i+1)
		}
		fmt.Fprintf(f, "\n")

		for i := range params.coords {
			for j := range params.coords[i] {
				fmt.Fprintf(f, "%.5f\t", params.coords[i][j])
			}
			fmt.Fprintf(f, "\n")
		}
		log.Println("Final total stress", initialStress)
		f.Close()
	}

}

func getTotalStress(similarityMatrix *core.DatasetSimilarities, coords []core.DatasetCoordinates) float64 {
	stress := 0.0
	for i := 0; i < len(coords); i++ {
		for j := i + 1; j < len(coords); j++ {
			actual := core.SimilarityToDistance(similarityMatrix.Get(i, j))
			measured := getDistance(coords[i], coords[j])
			if !math.IsInf(actual, 0) {
				stress += (actual - measured) * (actual - measured)
			}
		}
	}
	return math.Sqrt(stress)
}
