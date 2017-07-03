package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/giagiannis/data-profiler/core"
)

type expAccuracyParams struct {
	mlScript    *string         // script used for approximation
	output      *string         // output file path
	repetitions *int            // number of times to repeat experiment
	threads     *int            // number of threads to utilize
	datasets    []*core.Dataset //datasets to use

	coords    []core.DatasetCoordinates // coords of datasets
	evaluator core.DatasetEvaluator     // evaluator of the datasets

	samplingRates []float64 // samplings rates to run
}

func expAccuracyParseParams() *expAccuracyParams {
	params := new(expAccuracyParams)
	params.mlScript =
		flag.String("ml", "", "ML script to use for approximation")
	params.output =
		flag.String("o", "", "output path")
	params.repetitions =
		flag.Int("r", 1, "number of repetitions")
	params.threads =
		flag.Int("t", 1, "number of threads")
	loger :=
		flag.String("l", "", "log file")

	coordsFile :=
		flag.String("c", "", "coordinates file")
	scoresFile :=
		flag.String("s", "", "scores file")
	inputPath :=
		flag.String("i", "", "input path")
	srString :=
		flag.String("sr", "", "comma separated sampling rates")

	flag.Parse()
	setLogger(*loger)
	if *params.mlScript == "" || *params.output == "" || *coordsFile == "" ||
		*scoresFile == "" || *inputPath == "" || *srString == "" {
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// sampling rates parsing
	a := strings.Split(*srString, ",")
	params.samplingRates = make([]float64, 0)
	for i := range a {
		v, err := strconv.ParseFloat(a[i], 64)
		if err == nil {
			params.samplingRates = append(params.samplingRates, v)
		}
	}

	// datasets parsing
	params.datasets = core.DiscoverDatasets(*inputPath)

	// coordinates file parsing
	buf, err := ioutil.ReadFile(*coordsFile)
	if err != nil {
		log.Fatalln(err)
	}
	params.coords = core.DeserializeCoordinates(buf)

	// evaluator allocation
	params.evaluator, err = core.NewDatasetEvaluator(core.FileBasedEval, map[string]string{"scores": *scoresFile})
	if err != nil {
		log.Fatalln(err)
	}

	return params
}

func expAccuracyRun() {
	// inititializing steps
	params := expAccuracyParseParams()
	rand.Seed(int64(time.Now().Nanosecond()))
	output := setOutput(*params.output)
	defer output.Close()

	results := make(map[float64][]map[string]float64)
	for _, sr := range params.samplingRates {
		results[sr] = make([]map[string]float64, 0)
	}

	// threads configuration
	sync := make(chan bool, *params.threads)
	resChannel := make(chan resChannelResult)
	for i := 0; i < *params.threads; i++ {
		sync <- true
	}

	for r := 0; r < *params.repetitions; r++ {
		for _, sr := range params.samplingRates {
			modeler := core.NewModeler(params.datasets, sr, params.coords, params.evaluator)
			modeler.Configure(map[string]string{"script": *params.mlScript})
			go runModeler(sr, modeler, sync, resChannel)
		}
	}
	noResults := *params.repetitions * len(params.samplingRates)
	for i := 0; i < noResults; i++ {
		v := <-resChannel
		results[v.sr] = append(results[v.sr], v.res)
	}

	log.Println(results)
}

type resChannelResult struct {
	sr  float64
	res map[string]float64
}

func runModeler(sr float64, modeler core.Modeler, sync chan bool, resChannel chan resChannelResult) {
	<-sync
	err := modeler.Run()
	if err != nil {
		log.Println(err)
	}
	res := modeler.ErrorMetrics()
	res["execTime"] = modeler.ExecTime()
	res["evalTime"] = modeler.EvalTime()
	resChannel <- resChannelResult{sr, res}
	sync <- true
}
