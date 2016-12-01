package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/giagiannis/data-profiler/core"
)

type mdsParams struct {
	logfile *string // the logfile to be used
	output  *string // the output file

	script       *string                   // the mds script to run
	k            *int                      // the number of the coordinates to eval
	similarities *core.DatasetSimilarities // the similarity matrix
}

func mdsParseParams() *mdsParams {
	params := new(mdsParams)
	params.k =
		flag.Int("k", 2, "the number of the principal coordinates to use")
	params.script =
		flag.String("sc", "", "the script to be used for the MDS eval")
	similaritiesPath :=
		flag.String("sim", "", "the path of the similarity matrix file")
	params.logfile =
		flag.String("l", "", "the logfile to be used")
	params.output =
		flag.String("o", "", "the output file")
	flag.Parse()
	setLogger(*params.logfile)

	if *similaritiesPath == "" || *params.output == "" || *params.script == "" {
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	params.similarities = core.NewDatasetSimilarities(nil)
	log.Println("Reading", *similaritiesPath)
	f, err := os.Open(*similaritiesPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	err = params.similarities.Deserialize(b)

	if err != nil {
		log.Fatalln(err)
	}
	return params
}

func mdsRun() {
	params := mdsParseParams()
	mds := core.NewMDScaling(params.similarities, *params.k, *params.script)
	err := mds.Compute()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(mds.Coordinates())
}
