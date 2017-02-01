package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/giagiannis/data-profiler/core"
)

type expAccuracyParams struct {
	mlScript    *string // script used for approximation
	output      *string // output file path
	repetitions *int    // number of times to repeat experiment
	threads     *int    // number of threads to utilize

	coords []core.DatasetCoordinates // coords of datasets
	scores []float64                 // scores of datasets

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
	idxFile :=
		flag.String("i", "", "index file")
	srString :=
		flag.String("sr", "", "comma separated sampling rates")

	flag.Parse()
	setLogger(*loger)
	if *params.mlScript == "" || *params.output == "" || *coordsFile == "" ||
		*scoresFile == "" || *idxFile == "" || *srString == "" {
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

	// idx file parsing
	f, err := os.Open(*idxFile)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	idx := make([]string, 0)
	for i, line := range strings.Split(string(buf), "\n") {
		a := strings.Split(line, "\t")
		if len(a) == 2 {
			j, err := strconv.ParseInt(a[0], 10, 32)
			if err != nil || int(j) != i {
				log.Fatalln(err)
			}
			idx = append(idx, a[1])
		}
	}
	f.Close()

	// coordinates file parsing
	f, err = os.Open(*coordsFile)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err = ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	params.coords = make([]core.DatasetCoordinates, 0)
	for i, line := range strings.Split(string(buf), "\n") {
		a := strings.Split(line, " ")
		res := make(core.DatasetCoordinates, 0)
		if i > 0 && len(a) > 0 {
			for _, s := range a {
				if s != "" {
					v, err := strconv.ParseFloat(s, 64)
					if err != nil {
						log.Fatalln(err)
					}
					res = append(res, v)
				}
			}
			if len(res) > 0 {
				params.coords = append(params.coords, res)
			}
		}
	}
	f.Close()

	// scores
	f, err = os.Open(*scoresFile)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err = ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	scores := core.NewDatasetScores()
	scores.Deserialize(buf)
	params.scores = make([]float64, len(scores.Scores))
	for i, path := range idx {
		params.scores[i] = scores.Scores[path]
	}
	f.Close()

	return params
}

type evalAppxResults struct {
	mse, mape, mapeCorrected float64
}

func expAccuracyRun() {
	// inititializing steps
	params := expAccuracyParseParams()
	rand.Seed(int64(time.Now().Nanosecond()))
	output := setOutput(*params.output)
	fmt.Fprintln(output,
		"sr",
		"mse-avg", "mse-perc-0", "mse-perc-25", "mse-perc-50", "mse-perc-75", "mse-perc-100",
		"mape-avg", "mape-perc-0", "mape-perc-25", "mape-perc-50", "mape-perc-75", "mape-perc-100",
		"mapec-avg", "mapec-perc-0", "mapec-perc-25", "mapec-perc-50", "mapec-perc-75", "mapec-perc-100",
	)

	// create random permutation
	slice := make([]int, len(params.coords))
	for i := 0; i < len(slice); i++ {
		slice[i] = i
	}

	testset := generateSet(slice[0:int(float64(len(slice))*1.0)], params.coords, params.scores)

	executeScript := func(script, trainset, testset string) []float64 {
		cmd := exec.Command(script, trainset, testset)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(err)
		}
		result := make([]float64, 0)
		for _, line := range strings.Split(string(out), "\n") {
			val, err := strconv.ParseFloat(line, 64)
			if err == nil {
				result = append(result, val)
			}

		}
		return result
	}

	mse := func(predicted, actual []float64) float64 {
		if len(predicted) != len(actual) {
			log.Println("Predicted and actual values not of the same size")
		}
		sum := 0.0
		for i := range predicted {
			sum += (predicted[i] - actual[i]) * (predicted[i] - actual[i])
		}
		return sum / float64(len(predicted))
	}

	mape := func(predicted, actual []float64) float64 {
		if len(predicted) != len(actual) {
			log.Println("Predicted and actual values not of the same size")
		}
		sum := 0.0
		for i := range predicted {
			if actual[i] <= 0 {
				log.Println("Actual value not permitted", actual[i])
			} else {
				sum += math.Abs(float64(actual[i]-predicted[i]) / float64(actual[i]))
			}
		}
		return sum / float64(len(predicted))
	}

	mapeCorrected := func(predicted, actual []float64) float64 {
		if len(predicted) != len(actual) {
			log.Println("Predicted and actual values not of the same size")
		}
		sum := 0.0
		for i := range predicted {
			if actual[i] <= 0 || predicted[i] <= 0 {
				log.Println("Actual value not permitted", actual[i])
			} else {
				sum += math.Abs(math.Log(predicted[i]) - math.Log(actual[i]))
			}
		}
		return sum / float64(len(predicted))

	}

	eval := func(sr float64) *evalAppxResults {
		perm := rand.Perm(len(params.coords))
		trainsetIndexes := perm[0:int(float64(len(perm))*sr)]
		trainset := generateSet(trainsetIndexes, params.coords, params.scores)
		appxScores := executeScript(*params.mlScript, trainset, testset)
		res := new(evalAppxResults)
		res.mse = mse(appxScores, params.scores)
		res.mape = mape(appxScores, params.scores)
		res.mapeCorrected = mapeCorrected(appxScores, params.scores)
		os.Remove(trainset)
		return res
	}
	// execute
	for _, sr := range params.samplingRates {
		resultsMSE, resultsMAPE, resultsMAPECorrected :=
			make([]float64, 0), make([]float64, 0), make([]float64, 0)
		done := make(chan *evalAppxResults)
		slots := make(chan bool, *params.threads)
		for i := 0; i < *params.threads; i++ {
			slots <- true
		}

		for i := 0; i < *params.repetitions; i++ {
			go func(done chan *evalAppxResults, slots chan bool, repetition int) {
				log.Printf("[thread-%d] Starting calculation for SR %.2f\n", repetition, sr)
				<-slots
				done <- eval(sr)
				slots <- true
				log.Printf("[thread-%d] Done calculation for SR %.2f\n", repetition, sr)
			}(done, slots, i)
		}
		for i := 0; i < *params.repetitions; i++ {
			v := <-done
			resultsMSE = append(resultsMSE, v.mse)
			resultsMAPE = append(resultsMAPE, v.mape)
			resultsMAPECorrected = append(resultsMAPECorrected, v.mapeCorrected)
		}
		metricFormat := "%.5f\t%.5f\t%.5f\t%.5f\t%.5f\t%.5f"
		format := "%.5f"
		for i := 0; i < 3; i++ {
			format += "\t" + metricFormat
		}
		format += "\n"

		fmt.Fprintf(output, format,
			sr,
			getAverage(resultsMSE),
			getPercentile(resultsMSE, 0),
			getPercentile(resultsMSE, 25),
			getPercentile(resultsMSE, 50),
			getPercentile(resultsMSE, 75),
			getPercentile(resultsMSE, 100),
			getAverage(resultsMAPE),
			getPercentile(resultsMAPE, 0),
			getPercentile(resultsMAPE, 25),
			getPercentile(resultsMAPE, 50),
			getPercentile(resultsMAPE, 75),
			getPercentile(resultsMAPE, 100),
			getAverage(resultsMAPECorrected),
			getPercentile(resultsMAPECorrected, 0),
			getPercentile(resultsMAPECorrected, 25),
			getPercentile(resultsMAPECorrected, 50),
			getPercentile(resultsMAPECorrected, 75),
			getPercentile(resultsMAPECorrected, 100),
		)
	}
	os.Remove(testset)
}
