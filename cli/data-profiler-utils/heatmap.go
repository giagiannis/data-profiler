package main

import (
	"encoding/gob"
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
	scores       map[string]float64
	logfile      *string
}

func heatmapParseParams() *heatmapParams {
	params := new(heatmapParams)
	similaritiesPath :=
		flag.String("sm", "", "datasets similarities file")
	scoresPath :=
		flag.String("sc", "", "dataset scores file")
	params.logfile =
		flag.String("l", "", "logfile (default: stdout)")

	flag.Parse()
	if *similaritiesPath == "" ||
		*scoresPath == "" {
		fmt.Fprintf(os.Stderr,
			"Needed arguments not provided: type -h to see usage\n")
		os.Exit(1)
	}

	// reading similarities
	params.similarities = *core.NewDatasetSimilarities(nil)
	log.Println("Reading", *similaritiesPath)
	f, err := os.Open(*similaritiesPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = params.similarities.Deserialize(b)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	// reading scores
	log.Println("Reading", *scoresPath)
	f, err = os.Open(*scoresPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	d := gob.NewDecoder(f)
	err = d.Decode(&params.scores)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	return params
}

func heatmapRun() {
	params := heatmapParseParams()
	setLogger(*params.logfile)
	// order the scores by their
	list := sortScores(params.scores)
	for i := 0; i < len(list); i++ {
		for j := 0; j < len(list); j++ {
			fmt.Printf("%.5f ",
				params.similarities.Get(list[i].path, list[j].path))
		}
		fmt.Println()
	}
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
