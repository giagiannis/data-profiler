package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"

	"github.com/giagiannis/data-profiler/core"
)

type heatmapParams struct {
	similarities core.DatasetSimilarities
	scores       map[string]float64
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
		fmt.Fprintf(os.Stderr,
			"Needed arguments not provided: type -h to see usage\n")
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
	log.Println("Reading", *scoresPath)
	f, err = os.Open(*scoresPath)
	if err != nil {
		log.Fatalln(err)
	}
	d := gob.NewDecoder(f)
	err = d.Decode(&params.scores)
	if err != nil {
		log.Fatalln(err)
	}
	return params
}

func heatmapRun() {
	params := heatmapParseParams()
	list := sortScores(params.scores)
	log.Println("Creating temp file for heatmap data")
	fdata, err := ioutil.TempFile("/tmp", "heatmap-data-")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.Remove(fdata.Name())

	for i := 0; i < len(list); i++ {
		for j := 0; j < len(list); j++ {
			fmt.Fprintf(fdata, "%d %d %.5f\n",
				i, j,
				params.similarities.Get(list[i].path, list[j].path))
		}
		fmt.Fprintln(fdata)
	}
	fdata.Close()

	log.Println("Creating temp file for heatmap gnuplot script")
	fscript, err := ioutil.TempFile("/tmp", "heatmap-script-")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.Remove(fscript.Name())
	fmt.Fprintln(fscript, gnuplotScript1(fdata.Name(), *params.output, len(list)))
	fscript.Close()

	log.Println("Creating script")
	cmd := exec.Command("gnuplot", fscript.Name())
	_, err = cmd.Output()
	if err != nil {
		log.Fatalln(err)
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

func gnuplotScript1(input, output string, nrdatasets int) string {
	gnuplotScript := `inputFile='%s'
outputFile='%s.eps'
nrdatasets=%d
set title "Similarity Matrix heatmap"
set terminal postscript eps size 7,5.0 enhanced color font 'Arial,34'
set output outputFile
set xlabel "Dataset index"
set ylabel "Dataset index"
set xrange [-0.5:nrdatasets-0.5]
set yrange [-0.5:nrdatasets-0.5]
plot inputFile u 2:1:3 w image
system("epstopdf ".outputFile." && rm ".outputFile)`
	return fmt.Sprintf(gnuplotScript, input, output, nrdatasets)
}
