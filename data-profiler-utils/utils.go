package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

const (
	optsDelimiter = ","
	kvDelimiter   = "="
)

func parseOptions(optString string) map[string]string {
	opts := make(map[string]string)
	if optString == "" {
		return opts
	}
	for _, option := range strings.Split(optString, optsDelimiter) {
		sp := strings.Split(option, kvDelimiter)
		if len(sp) != 2 {
			fmt.Fprintln(os.Stderr, "Malformed options")
			os.Exit(1)
		}
		opts[sp[0]] = sp[1]
	}
	return opts
}

func setLogger(logfile string) {
	if logfile != "" {
		f, er := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if er != nil {
			fmt.Println(er)
			os.Exit(1)
		} else {
			log.SetOutput(f)
		}
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

}

func setOutput(output string) *os.File {
	outF := os.Stdout
	if output != "" {
		var err error
		outF, err = os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return outF

}

// returns the average of a float slice
func getAverage(sl []float64) float64 {
	if sl == nil || len(sl) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range sl {
		sum += v
	}
	return sum / float64(len(sl))
}

// returns the n-th percentile of a float slice (50 for median)
func getPercentile(sl []float64, percentile int) float64 {
	sort.Float64s(sl)
	idx := int((float64(percentile) / 100.0) * float64(len(sl)))
	if idx >= len(sl) {
		idx = len(sl) - 1
	}
	return sl[idx]
}
