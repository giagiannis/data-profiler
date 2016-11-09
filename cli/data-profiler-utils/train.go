package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/giagiannis/data-profiler/core"
)

type trainParams struct {
	testset     *core.Dataset   // path of the test set
	script      *string         // path of the ML script
	datasets    []*core.Dataset // path of the input datasets
	concurrency *int            // number of threads to be utilized
	output      *string         // path of the outfile
	logfile     *string         // logfile path
}

func trainParseParams() *trainParams {
	res := new(trainParams)
	testset :=
		flag.String("t", "", "path of the test set")
	res.script =
		flag.String("s", "", "path of the ML script")
	input :=
		flag.String("i", "", "path of the datasets dir")
	res.output =
		flag.String("o", "", "path of the output file")
	res.logfile =
		flag.String("l", "", "path of the log file")
	res.concurrency =
		flag.Int("p", 1, "number of threads")
	flag.Parse()

	if *testset == "" ||
		*res.script == "" ||
		*res.output == "" ||
		*input == "" {
		fmt.Fprintf(os.Stderr,
			"Needed arguments not provided: type -h to see usage\n")
		os.Exit(1)
	}
	res.datasets = core.DiscoverDatasets(*input)
	res.testset = core.NewDataset(*testset)
	return res
}

func trainRun() {
	params := trainParseParams()
	setLogger(*params.logfile)
	results := make(map[string]float64)
	done := make(chan bool)
	c := make(chan bool, *params.concurrency)
	for i := 0; i < *params.concurrency; i++ {
		c <- true
	}
	lock := new(sync.Mutex)
	for _, d := range params.datasets {
		go func(d core.Dataset, c, done chan bool) {
			<-c
			log.Println("Training with ", d.Path())
			val, err := execMLScript(*params.script, d, *params.testset)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
			lock.Lock()
			results[d.Path()] = val
			lock.Unlock()
			c <- true
			done <- true
		}(*d, c, done)
	}
	for _ = range params.datasets {
		<-done
	}
	log.Println("Serializing output to ", *params.output)
	outfile, er := os.OpenFile(*params.output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer outfile.Close()

	if er != nil {
		fmt.Fprintln(os.Stderr, er)
		os.Exit(1)
	}
	e := gob.NewEncoder(outfile)
	err := e.Encode(results)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// function used to execute an ML script and returns its error
func execMLScript(scriptPath string, trainset, testset core.Dataset) (float64, error) {
	cmd := exec.Command(scriptPath, trainset.Path(), testset.Path())
	o, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		return -1, err
	}
	val, err := strconv.ParseFloat(string(o), 64)
	if err != nil {
		log.Println(err)
		return -1, err
	}
	return val, nil
}
