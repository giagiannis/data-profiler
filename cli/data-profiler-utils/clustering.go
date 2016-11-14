package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/giagiannis/data-profiler/core"
)

type clusteringParams struct {
	similarities *core.DatasetSimilarities // similarities
	scores       *core.DatasetScores       // scores file
	logfile      *string                   // log file
	concurrency  *int                      // max number of threads
	output       *string                   // output file
}

func clusteringParseParams() *clusteringParams {
	params := new(clusteringParams)

	similaritiesFile :=
		flag.String("si", "", "similarities file")
	scoresFile :=
		flag.String("sc", "", "scores file")
	params.logfile =
		flag.String("l", "", "logfile (default: stdout)")
	params.concurrency =
		flag.Int("t", 1, "number of threads")
	params.output =
		flag.String("o", "", "output file (default: stdout)")
	flag.Parse()
	setLogger(*params.logfile)

	if *similaritiesFile == "" || *scoresFile == "" {
		fmt.Fprintln(os.Stderr,
			"I need both similarities file and scores file")
		os.Exit(1)
	}

	// parse similarities file
	f, err := os.Open(*similaritiesFile)
	defer f.Close()
	if err != nil {
		log.Fatalln(err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	params.similarities = core.NewDatasetSimilarities(nil)
	err = params.similarities.Deserialize(buf)
	if err != nil {
		log.Fatalln(err)
	}

	// parse scores file
	fs, err := os.Open(*scoresFile)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err = ioutil.ReadAll(fs)
	if err != nil {
		log.Fatalln(err)
	}
	params.scores = core.NewDatasetScores()
	err = params.scores.Deserialize(buf)
	if err != nil {
		log.Fatalln(err)
	}

	return params
}

func clusteringRun() {
	params := clusteringParseParams()
	log.Println("Initializing clustering object")
	cls := core.NewClustering(params.similarities)
	cls.SetConcurrency(*params.concurrency)
	log.Println("Executing computation")
	cls.Compute()
	log.Println("Done")
	outF := os.Stdout
	if *params.output != "" {
		var err error
		outF, err = os.OpenFile(*params.output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		defer outF.Close()
	}
	maxTreeHeight, _ := cls.Results().Heights()
	fmt.Fprintf(outF, "level num_of_clusters avg_error avg_max_min_error\n")
	for i := 0; i <= maxTreeHeight; i++ {
		avg, diff := evaluateClusters(cls.Results().GetClusters(i), params.scores)
		fmt.Fprintf(outF, "%d %d %.5f %.5f\n", i,
			len(cls.Results().GetClusters(i)), avg, diff)
	}
	if outF != os.Stdout {
		log.Println("Results written to", *params.output)
	}
}

func evaluateClusters(clusters [][]*core.Dataset, scores *core.DatasetScores) (float64, float64) {
	sumAvgs, sumDiffs := 0.0, 0.0
	for _, c := range clusters {
		s, d := evaluateCluster(c, scores)
		sumAvgs += s
		sumDiffs += d
	}
	return sumAvgs / float64(len(clusters)), sumDiffs / float64(len(clusters))
}

// returns various metrics for the specified clusters
func evaluateCluster(datasets []*core.Dataset, scores *core.DatasetScores) (float64, float64) {
	sum := 0.0
	maxError, minError := -1.0, -1.0
	for _, d := range datasets {
		s := scores.Scores[d.Path()]
		sum += s
		if maxError == -1 || s > maxError {
			maxError = s
		}
		if minError == -1 || s < minError {
			minError = s
		}
	}
	return sum / float64(len(datasets)), maxError - minError
}
