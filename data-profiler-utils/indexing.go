package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/giagiannis/data-profiler/core"
)

type indexingParams struct {
	estimator   core.DatasetSimilarityEstimator // estimator object
	coordinates []core.DatasetCoordinates       // coordinates of datasets
	datasets    []*core.Dataset                 // new datasets to index
	script      *string                         // indexing script needed by the indexer
	output      *string                         // where to store the output
	logfile     *string                         // logfile of execution
	k           *int                            // determines the number of datasets to compare
	concurrency *int                            //number of threads to be utilized for the exec
}

func indexingParseParams() *indexingParams {
	params := new(indexingParams)
	estimatorFile :=
		flag.String("e", "", "the serialized estimator file")
	coordinatesFile :=
		flag.String("c", "", "the dataset coordinates file")
	params.script =
		flag.String("s", "", "the script file to be used by the indexer")
	params.output =
		flag.String("o", "", "the output file")
	params.logfile =
		flag.String("l", "", "the log file")
	datasetsPath :=
		flag.String("i", "", "the new datasets to be indexed")
	params.k =
		flag.Int("k", 0, "the number of datasets to compare for indexing - the default is equal to the number of datasets in the matrix")
	params.concurrency =
		flag.Int("t", 1, "the number of threads to spawn for the indexing")
	flag.Parse()
	setLogger(*params.logfile)

	if *estimatorFile == "" || *coordinatesFile == "" || *params.script == "" ||
		*params.output == "" || *datasetsPath == "" {
		fmt.Fprintln(os.Stderr, "Missing arguments, usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// parse estimator object
	f, err := os.Open(*estimatorFile)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	buffer, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	params.estimator = core.DeserializeSimilarityEstimator(buffer)

	// parse coordinates file
	f, err = os.Open(*coordinatesFile)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	coordBuffer, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	params.coordinates = make([]core.DatasetCoordinates, 0)
	for _, l := range strings.Split(string(coordBuffer), "\n") {
		e := false
		tuple := make(core.DatasetCoordinates, 0)
		for _, v := range strings.Split(l, " ") {
			if !e {
				v = strings.TrimSpace(v)
				if v != "" {
					val, err := strconv.ParseFloat(v, 64)
					e = (err != nil)
					tuple = append(tuple, val)
				}
			}
		}
		if !e && len(tuple) > 0 {
			params.coordinates = append(params.coordinates, tuple)
		}
	}

	// parse datasets
	params.datasets = core.DiscoverDatasets(*datasetsPath)

	// parse k
	if *params.k == 0 {
		*params.k = len(params.estimator.Datasets())
	}

	return params
}

func indexingRun() {
	params := indexingParseParams()
	indexer := core.NewOnlineIndexer(params.estimator, params.coordinates, *params.script)
	indexer.DatasetsToCompare(*params.k)
	c := make(chan bool, *params.concurrency)
	done := make(chan resultsStruct)
	for j := 0; j < *params.concurrency; j++ {
		c <- true
	}
	for i, d := range params.datasets {
		go func(c chan bool, done chan resultsStruct, d *core.Dataset, i int) {
			<-c
			start := time.Now()
			coo, str, err := indexer.Calculate(d)
			duration := time.Since(start) / 1000
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			c <- true
			done <- resultsStruct{i, coo, str, duration}
		}(c, done, d, i)
	}
	for range params.datasets {
		res := <-done
		fmt.Println(res)
	}
}

type resultsStruct struct {
	id       int
	coords   core.DatasetCoordinates
	stress   float64
	duration time.Duration
}
