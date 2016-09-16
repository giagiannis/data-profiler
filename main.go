package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/giagiannis/data-profiler/analysis"
	"github.com/giagiannis/data-profiler/optimization"

	"gopkg.in/gcfg.v1"
)

var confFile *string

type Configuration struct {
	Datasets struct {
		Training *string
		Testing  *string
		Copies   *int
	}
	Scripts struct {
		Ml       *string
		Analysis *string
	}

	SA struct {
		Iterations *int
		TempInit   *float64
		TempDecay  *float64
	}

	Other struct {
		Threads       *int
		Optimizations *string
	}
}

func (c Configuration) String() string {
	buffer := ""
	buffer += fmt.Sprintf(
		"(datasets.training, %s) (datasets.testing, %s) (datasets.copies, %d) "+
			"(sa.iterations, %d) (sa.tempInit, %.5f) (sa.tempDecay, %.5f) "+
			"(scripts.Ml, %s) (scripts.Analysis, %s) "+
			"(other.threads, %d) (other.optimizations, %b)",
		*c.Datasets.Training, *c.Datasets.Testing, *c.Datasets.Copies,
		*c.SA.Iterations, *c.SA.TempInit, *c.SA.TempDecay,
		*c.Scripts.Ml, *c.Scripts.Analysis,
		*c.Other.Threads, *c.Other.Optimizations)
	return buffer
}

func (c *Configuration) ApplyDiffs(other Configuration) {
	if *other.Datasets.Training != "" {
		c.Datasets.Training = other.Datasets.Training
	}
	if *other.Datasets.Testing != "" {
		c.Datasets.Testing = other.Datasets.Testing
	}
	if *other.Datasets.Copies != -1 {
		c.Datasets.Copies = other.Datasets.Copies
	}
	if *other.Other.Optimizations != "" {
		c.Other.Optimizations = other.Other.Optimizations
	}
	if *other.Other.Threads != -1 {
		c.Other.Threads = other.Other.Threads
	}
	if *other.SA.Iterations != -1 {
		c.SA.Iterations = other.SA.Iterations
	}
	if *other.SA.TempDecay != -1.0 {
		c.SA.TempDecay = other.SA.TempDecay
	}
	if *other.SA.TempInit != -1 {
		c.SA.TempInit = other.SA.TempInit
	}
	if *other.Scripts.Analysis != "" {
		c.Scripts.Analysis = other.Scripts.Analysis
	}
	if *other.Scripts.Ml != "" {
		c.Scripts.Ml = other.Scripts.Ml
	}
}

func parseParams() (*string, *Configuration) {
	confFile := flag.String("conf", "", "the configuration file")

	cliConf := new(Configuration)
	cliConf.Datasets.Training =
		flag.String("datasets.training", "", "The path of the dataset to be analyzed")
	cliConf.Datasets.Testing =
		flag.String("datasets.testing", "", "The test set used to extract the classification score")
	cliConf.Datasets.Copies =
		flag.Int("datasets.copies", -1, "Number of subsets to break the original dataset")
	cliConf.Scripts.Analysis =
		flag.String("scripts.analysis", "", "The analysis script to be executed")
	cliConf.Scripts.Ml =
		flag.String("scripts.ml", "", "The ML script to be executed")
	cliConf.Other.Optimizations =
		flag.String("other.optimizations", "", "Sets optimizations")
	cliConf.Other.Threads =
		flag.Int("other.threads", -1, "Number of threads to spawn for parallel tasks")
	cliConf.SA.Iterations =
		flag.Int("sa.iterations", -1, "Max iterations for the optimizer")
	cliConf.SA.TempDecay =
		flag.Float64("sa.tempDecay", -1, "Temperature decay")
	cliConf.SA.TempInit =
		flag.Float64("sa.tempInit", -1, "Initial temperature")
	flag.Parse()
	return confFile, cliConf
}

func main() {
	confFile, cliCfg := parseParams()
	if *confFile == "" {
		fmt.Println("Error in argument parsing, please provide conf file or see help")
		os.Exit(1)
	}

	cfg := new(Configuration)
	gcfg.ReadFileInto(cfg, *confFile)
	cfg.ApplyDiffs(*cliCfg)

	datasets := analysis.DatasetPartition(
		*analysis.NewDataset(*cfg.Datasets.Training),
		*cfg.Datasets.Copies)
	m := analysis.NewManager(datasets, *cfg.Other.Threads, *cfg.Scripts.Analysis)
	m.Analyze()
	if *cfg.Other.Optimizations == "true" {
		m.OptimizationResultsPruning()
	}
	optimizer := optimization.NewSimulatedAnnealingOptimizer(
		*cfg.Scripts.Ml,
		*analysis.NewDataset(*cfg.Datasets.Testing),
		*cfg.SA.Iterations,
		*cfg.SA.TempDecay,
		*cfg.SA.TempInit,
		m.Results())
	optimizer.Run()
	// cleanup before exiting
	analysis.DatasetPartitionCleanup(datasets)
}
