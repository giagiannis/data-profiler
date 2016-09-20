package analysis

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ScriptAnalyzer implements the Analyzer interface and executes any exec script
// in order to run the analysis
type ScriptAnalyzer struct {
	dataset  Dataset    // dataset object
	started  time.Time  // timestamp when analysis started
	finished time.Time  // timestamp when analysis finished
	status   Status     // status of the analyzer
	result   Result     // the result of the analysis
	mutex    sync.Mutex // Mutex used to protect the status variable
	script   string     // path of the R script to execute for analysis
}

// NewScriptAnalyzer is a constructor for ScriptAnalyzer
// returns a reference to the newly allocated object
func NewScriptAnalyzer(dataset Dataset, script string) *ScriptAnalyzer {
	return &ScriptAnalyzer{dataset, time.Time{}, time.Time{}, PENDING, nil, sync.Mutex{}, script}
}

// Dataset is a getter for the dataset object
func (a *ScriptAnalyzer) Dataset() Dataset {
	return a.dataset
}

// Analyze method for the ScriptAnalyzer. False is returned if the analysis fails
func (a *ScriptAnalyzer) Analyze() bool {
	a.mutex.Lock()
	a.status = ANALYZING
	a.started = time.Now()
	a.mutex.Unlock()
	cmd := exec.Command(a.script, a.dataset.Path())
	out, er := cmd.Output()
	if er != nil { // if the command is not executed, return false
		return false
	}

	// parse command's output
	outArray := strings.Split(string(out), "\n")
	nComponents, e := strconv.Atoi(outArray[0])
	if e != nil {
		return false
	}

	a.result = make([]float64, nComponents)

	for k, v := range strings.Split(strings.TrimSpace(outArray[1]), " ") {
		a.result[k], e = strconv.ParseFloat(v, 64)
		if e != nil {
			return false
		}
	}

	a.mutex.Lock()
	a.status = ANALYZED
	a.finished = time.Now()
	a.mutex.Unlock()
	return true
}

// Result function is used to fetch the results of the analysis.
func (a *ScriptAnalyzer) Result() Result {
	return a.result
}

// Status returns the status of the Analyzer
func (a *ScriptAnalyzer) Status() Status {
	a.mutex.Lock()
	res := a.status
	a.mutex.Unlock()
	return res
}

// String method override
func (a ScriptAnalyzer) String() string {
	return "(" + a.dataset.Id() + ")"
}

// Duration returns the elapsed time for the analysis
func (a *ScriptAnalyzer) Duration() float64 {
	return a.finished.Sub(a.started).Seconds()
}
