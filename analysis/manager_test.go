package analysis

import (
	"fmt"
	"os"
	"testing"
)

const (
	TESTSET         = "../datasets/shuttle-test.csv"
	TRAINSET        = "../datasets/shuttle-train.csv"
	ANALYSIS_SCRIPT = "../r_scripts/pca.R"
)

// Function used to return an array of datasets based on their filenames
// executed in parallel
func createDatasets(fileNames []string) []Dataset {
	datasets := make([]Dataset, len(fileNames))
	done := make(chan bool, 1)
	cookies := make(chan bool, len(fileNames)+1)
	for i := 0; i < 8; i++ {
		cookies <- true
	}
	for i, n := range fileNames {
		go func(i int, n string, cookies chan bool) {
			<-cookies
			datasets[i] = *NewDataset(n)
			cookies <- true
			done <- true
		}(i, n, cookies)
	}

	for i := 0; i < len(fileNames); i++ {
		<-done
	}
	return datasets
}

func TestManagerAnalyze(t *testing.T) {
	rows, cols, d := 1000, 3, 20
	fileNames := make([]string, d)
	for i := 0; i < d; i++ {
		fileNames[i] = createRandomDataset(rows, cols)
	}
	fmt.Println("Created datasets")
	datasets := createDatasets(fileNames)
	m := NewManager(datasets, 8, "../r_scripts/pca.R")
	m.Analyze()

	for _, f := range datasets {
		if _, ok := m.results[f]; !ok {
			t.Log(f.Id(),
				" not analyzed:",
				m.results)
			t.Fail()
		} else if len(m.results[f]) != cols*cols {
			t.Log("Serialized results missing from ",
				f.Id(),
				": ",
				m.results[f])
			t.Fail()
		}
	}

	for _, f := range fileNames {
		os.Remove(f)
	}
}

func TestManagetOptimizationResultsPruning(t *testing.T) {
	datasets := DatasetPartition(*NewDataset(TRAINSET), 100)
	manager := NewManager(datasets, 8, ANALYSIS_SCRIPT)
	manager.Analyze()

	manager.OptimizationResultsPruning()

}
