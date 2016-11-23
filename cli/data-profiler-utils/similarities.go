package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/giagiannis/data-profiler/core"
)

type similaritiesParams struct {
	input   *string                              //datasets directory to be discovered
	output  *string                              // output file
	simType *core.DatasetSimilarityEstimatorType // similarity type
	logfile *string                              // logfile
	options *string                              // options for the estimators
}

func similaritiesParseParams() *similaritiesParams {
	params := new(similaritiesParams)
	params.input =
		flag.String("i", "", "input datasets path")
	params.output =
		flag.String("o", "", "where to store similarities file")
	estType :=
		flag.String("t", "BHATTACHARYYA", "similarity type [JACOBBI|BHATTACHARYYA]")
	params.logfile =
		flag.String("l", "", "logfile (default: stderr)")
	params.options =
		flag.String("opt", "", "options in the form val1=key1,val2=key2 (list for opts list)")
	flag.Parse()
	setLogger(*params.logfile)

	if *estType == "BHATTACHARYYA" {
		params.simType = new(core.DatasetSimilarityEstimatorType)
		*params.simType = core.BHATTACHARYYA
	} else if *estType == "JACOBBI" {
		params.simType = new(core.DatasetSimilarityEstimatorType)
		*params.simType = core.JACOBBI
	} else if *estType == "SCRIPT" {
		params.simType = new(core.DatasetSimilarityEstimatorType)
		*params.simType = core.SCRIPT
	}

	if *params.options == "list" {
		similarityTypes := []core.DatasetSimilarityEstimatorType{
			core.JACOBBI, core.BHATTACHARYYA, core.SCRIPT,
		}
		for i, v := range similarityTypes {
			fmt.Println(i+1, v)
			a := core.NewDatasetSimilarityEstimator(v, nil)
			for k, v := range a.Options() {
				fmt.Println("\t", k, ":", v)
			}
		}
		os.Exit(0)
	}

	if *params.input == "" ||
		*params.output == "" ||
		params.simType == nil {
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	return params
}

func similaritiesRun() {
	params := similaritiesParseParams()
	datasets := core.DiscoverDatasets(*params.input)
	est := core.NewDatasetSimilarityEstimator(*params.simType, datasets)
	est.Configure(parseOptions(*params.options))
	est.Compute()

	outfile, er := os.OpenFile(*params.output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if er != nil {
		fmt.Fprintln(os.Stderr, er)
		os.Exit(1)
	}
	defer outfile.Close()
	outfile.Write(est.GetSimilarities().Serialize())
	log.Println("\n" + est.GetSimilarities().String())
}
