package core

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestRAnalyze(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	dataset := createPoolBasedDatasets(1000, 1, 5)[0]
	rAnalyzer := NewScriptAnalyzer(dataset, ANALYSIS_SCRIPT)

	ok := rAnalyzer.Analyze()

	if ok != true {
		t.Log("Analysis failed")
		t.Fail()
	}
	if rAnalyzer == nil {
		t.Log("Eigenvalues are null")
		t.Fail()
	}

	os.Remove(dataset.Path())
}

func TestRAnalyzerStatus(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	dataset := createPoolBasedDatasets(1000, 1, 5)[0]
	rAnalyzer := NewScriptAnalyzer(dataset, ANALYSIS_SCRIPT)
	if rAnalyzer.Status() != PENDING {
		t.Log("Status should be pending")
		t.Fail()
	}
	rAnalyzer.Analyze()
	if rAnalyzer.Status() != ANALYZED {
		t.Log("Error status")
		t.Fail()
	}
	os.Remove(dataset.Path())
}
