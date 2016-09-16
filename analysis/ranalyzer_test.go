package analysis

import "testing"

import "math/rand"
import "fmt"
import "time"
import "os"

// Function used to generate a random dataset - to be used for testing
// returns the path of the created dataset
func createRandomDataset(rows int, col int) string {
	t := time.Now()
	rand.Seed(int64(t.Nanosecond()))
	buffer := make([]byte, 5)
	rand.Read(buffer)
	filename := "/tmp/dataset-" + fmt.Sprintf("%x", buffer) + ".dat"
	file_content := ""
	for j := 0; j < col; j++ {
		file_content += fmt.Sprintf("%d", j)
		if j != col-1 {
			file_content += ","
		}
	}
	file_content += fmt.Sprintf("\n")
	for i := 0; i < rows; i++ {
		for j := 0; j < col; j++ {
			file_content += fmt.Sprintf("%.5f", rand.Float64())
			if j != col-1 {
				file_content += ","
			}
		}
		file_content += fmt.Sprintf("\n")
	}
	buffer = []byte(file_content)
	f, e := os.Create(filename)
	if e != nil {
		return ""
	}
	f.Write(buffer)
	f.Sync()
	f.Close()
	return filename
}

func TestRAnalyze(t *testing.T) {
	filename := createRandomDataset(100, 4)
	rAnalyzer := NewRAnalyzer(*NewDataset(filename), ANALYSIS_SCRIPT)

	ok := rAnalyzer.Analyze()

	if ok != true {
		t.Log("Analysis failed")
		t.Fail()
	}
	if rAnalyzer == nil {
		t.Log("Eigenvalues are null")
		t.Fail()
	}

	os.Remove(filename)
}

func TestRAnalyzerStatus(t *testing.T) {
	filename := createRandomDataset(500, 3)
	rAnalyzer := NewRAnalyzer(*NewDataset(filename), ANALYSIS_SCRIPT)
	if rAnalyzer.Status() != PENDING {
		t.Log("Status should be pending")
		t.Fail()
	}
	rAnalyzer.Analyze()
	if rAnalyzer.Status() != ANALYZED {
		t.Log("Error status")
		t.Fail()
	}
	os.Remove(filename)
}
