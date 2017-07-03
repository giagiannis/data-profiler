package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sort"
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
		if _, ok := results[v.sr]; !ok {
			results[v.sr] = make([]map[string]float64, 0)
		}
		results[v.sr] = append(results[v.sr], v.res)
	}
	log.Println(results)

	keys := writeResults(output, results, params.samplingRates)
	fmt.Println("Column names/indices:")
	fmt.Printf("%d - %s\n", 1, "sr")
	for i, k := range keys {
		fmt.Printf("%d - %s\n", i+2, k)
	}

}

// writeResults writes the results to the output file and returns a string slice
// containing the names of the CSV's columns
func writeResults(output *os.File, results map[float64][]map[string]float64, samplingRates []float64) []string {
	keys, keysFinal := make([]string, 0), make([]string, 0)
	getValue := func(key string, results []map[string]float64) []float64 {
		res := make([]float64, 0)
		for _, v := range results {
			res = append(res, v[key])
		}
		return res
	}
	for _, sr := range samplingRates {
		rLine := results[sr]
		if len(keys) == 0 { // get and print header
			for k := range rLine[0] {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			fmt.Fprintf(output, "sr")
			for _, k := range keys {
				for _, k2 := range []string{"mean", "stddev", "median"} {
					fmt.Fprintf(output, "\t%s", k+"-"+k2)
					keysFinal = append(keysFinal, k+"-"+k2)
				}
			}
			fmt.Fprintf(output, "\n")
		}

		fmt.Fprintf(output, "%.2f", sr)
		for _, k := range keys {
			values := getValue(k, rLine)
			mean, stddev, median := core.Mean(values), core.StdDev(values), core.Percentile(values, 50)
			fmt.Fprintf(output, "\t%.5f\t%.5f\t%.5f", mean, stddev, median)
		}
		fmt.Fprintf(output, "\n")
	}
	return keysFinal
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
		sync <- true
		return
	}
	res := modeler.ErrorMetrics()
	res["TimeExec"] = modeler.ExecTime()
	res["TimeEval"] = modeler.EvalTime()
	resChannel <- resChannelResult{sr, res}
	sync <- true
}
