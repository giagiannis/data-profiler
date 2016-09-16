package analysis

import "testing"

const DATASET = "../ml_scripts/shuttle-train.csv"

func TestDatasetPartition(t *testing.T) {
	datasets := DatasetPartition(*NewDataset(DATASET), 13)
	if datasets == nil {
		t.Log("Nil returned")
		t.FailNow()
	}

	for _, d := range datasets {
		if d.Id() == "" {
			t.Log("Nil returned")
			t.FailNow()
		}
	}
}
