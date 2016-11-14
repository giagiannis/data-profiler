package core

import (
	"log"
	"testing"
)

func TestDatasetEvaluator(t *testing.T) {
	eval := NewDatasetEvaluator(ML_SCRIPT, TESTSET)
	val, err := eval.Evaluate(TRAINSET)
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if val < 0 {
		log.Println("Wrong conversion")
		t.FailNow()
	}
}
