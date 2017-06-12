package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Modeler is the interface for the objects that model the dataset space.
type Modeler interface {
	//  Configure is responsible to provide the necessary configuration
	// options to the Modeler struct. Call it before Run.
	Configure(map[string]string) error
	// Run initiates the modeling process.
	Run() error

	// Datasets returns the datasets slice
	Datasets() []*Dataset
	// Samples returns the indices of the chosen datasets.
	Samples() []struct {
		idx int
		val float64
	}
	// AppxValues returns a slice of the approximated values
	AppxValues() []float64
}

// AbstractModeler implements the common methods of the Modeler structs
type AbstractModeler struct {
	datasets    []*Dataset           // the datasets the modeler refers to
	evaluator   DatasetEvaluator     // the evaluator struct that gets the values
	coordinates []DatasetCoordinates // the dataset coordinates

	samplingRate float64 // the portion of the datasets to examine
	// the dataset indices chosen for samples
	samples []struct {
		idx int
		val float64
	}
	appxValues []float64 // the appx values of ALL the datasets
}

// Datasets returns the datasets slice
func (a *AbstractModeler) Datasets() []*Dataset {
	return a.datasets
}

// Samples return the indices of the chosen datasets
func (a *AbstractModeler) Samples() []struct {
	idx int
	val float64
} {
	return a.samples
}

// AppxValues returns the values of all the datasets
func (a *AbstractModeler) AppxValues() []float64 {
	return a.appxValues
}

// ScriptBasedModeler utilizes a script to train an ML model and obtain is values
type ScriptBasedModeler struct {
	AbstractModeler
	script string // the script to use for modeling
}

// Configure expects the necessary conf options for the specified struct.
// Specifically, the following parameters are necessary:
// - script: the path of the script to use
func (m *ScriptBasedModeler) Configure(conf map[string]string) error {
	if val, ok := conf["script"]; ok {
		m.script = val
	} else {
		log.Println("script parameter is missing")
		return errors.New("script parameter is missing")
	}
	return nil
}

// Run executes the modeling process and populates the samples, realValues and
// appxValues slices.
func (m *ScriptBasedModeler) Run() error {
	// sample the datasets
	permutation := rand.Perm(len(m.datasets))
	s := int(math.Floor(m.samplingRate * float64(len(m.datasets))))

	// deploy samples
	var trainingSet, testSet [][]float64
	for _, idx := range permutation[:s] {
		val, err := m.evaluator.Evaluate(m.datasets[idx].Path())
		if err != nil {
			return err
		}
		m.samples = append(m.samples, struct {
			idx int
			val float64
		}{idx, val})
		trainingSet = append(trainingSet, append(m.coordinates[idx], val))
	}
	trainFile := createCSVFile(trainingSet, true)
	for _, v := range m.coordinates {
		testSet = append(testSet, v)
	}
	testFile := createCSVFile(testSet, false)
	appx, err := m.executeMLScript(trainFile, testFile)
	if err != nil {
		return err
	}
	m.appxValues = appx
	os.Remove(trainFile)
	os.Remove(testFile)
	return nil
}

// executeMLScript executes the ML script, utilizing the selected samples (indices)
// and populates the real and appx values slices
func (m *ScriptBasedModeler) executeMLScript(trainFile, testFile string) ([]float64, error) {
	var result []float64
	cmd := exec.Command(m.script, trainFile, testFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	outputString := string(out)
	array := strings.Split(outputString, "\n")
	result = make([]float64, len(m.datasets))
	for i := 0; i < len(m.datasets); i++ {
		val, err := strconv.ParseFloat(array[i], 64)
		if err != nil {
			log.Println(err)
		} else {
			result[i] = val
		}
	}
	return result, nil
}

// createCSVFile serializes a double float slice to a CSV file and returns
// the filename
func createCSVFile(matrix [][]float64, output bool) string {
	f, err := ioutil.TempFile("/tmp", "csv")
	if err != nil {
		log.Println(err)
	}
	cols := 0
	if len(matrix) > 0 {
		cols = len(matrix[0])
	}
	if output {
		cols--
	}

	for i := 1; i < cols+1; i++ {
		fmt.Fprintf(f, "x%d", i)
		if i < cols {
			fmt.Fprintf(f, ",")
		}
	}
	if output {
		fmt.Fprintf(f, ",class")
	}
	fmt.Fprintf(f, "\n")

	for i := range matrix {
		for j := range matrix[i] {
			fmt.Fprintf(f, "%.5f", matrix[i][j])
			if j < len(matrix[i])-1 {
				fmt.Fprintf(f, ",")
			}
		}
		fmt.Fprintf(f, "\n")
	}
	f.Close()
	return f.Name()
}
