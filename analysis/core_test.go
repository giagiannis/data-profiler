package analysis

import "testing"

func TestDatasetPartition(t *testing.T) {
	datasets := DatasetPartition(*NewDataset(TRAINSET), 13)
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
