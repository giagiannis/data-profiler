package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/giagiannis/data-profiler/core"
)

type heatmapParams struct {
	similarities core.DatasetSimilarities
	scores       core.DatasetScores
	logfile      *string
	output       *string
}

func heatmapParseParams() *heatmapParams {
	params := new(heatmapParams)
	similaritiesPath :=
		flag.String("sm", "", "datasets similarities file")
	scoresPath :=
		flag.String("sc", "", "dataset scores file")
	params.logfile =
		flag.String("l", "", "logfile (default: stdout)")
	params.output =
		flag.String("o", "", "output file")

	flag.Parse()
	setLogger(*params.logfile)
	if *similaritiesPath == "" ||
		*params.output == "" ||
		*scoresPath == "" {
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// reading similarities
	params.similarities = *core.NewDatasetSimilarities(nil)
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
	// reading scores
	params.scores = *core.NewDatasetScores()
	log.Println("Reading", *scoresPath)
	f, err = os.Open(*scoresPath)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err := ioutil.ReadAll(f)
	params.scores.Deserialize(buf)
	if err != nil {
		log.Fatalln(err)
	}
	return params
}

func heatmapRun() {
	params := heatmapParseParams()
	list := sortScores(params.scores.Scores)
	log.Println("Creating file for heatmap data")
	outfile, err := os.OpenFile(*params.output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer outfile.Close()

	for i := 0; i < len(list); i++ {
		for j := 0; j < len(list); j++ {
			fmt.Fprintf(outfile, "%d %d %.5f\n",
				i, j,
				params.similarities.Get(list[i].path, list[j].path))
		}
		fmt.Fprintln(outfile)
	}
	outfile.Close()
}

type ScoresPair struct {
	path  string
	score float64
}

type ScoresPairList []ScoresPair

func (s ScoresPairList) Len() int {
	return len(s)
}

func (s ScoresPairList) Less(i, j int) bool {
	return s[i].score < s[j].score
}

func (s ScoresPairList) Swap(i, j int) {
	t := s[i]
	s[i] = s[j]
	s[j] = t
}
func sortScores(scores map[string]float64) ScoresPairList {

	var list ScoresPairList
	for d, v := range scores {
		list = append(list, ScoresPair{d, v})

	}
	sort.Sort(list)
	return list
}
