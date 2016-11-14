package core

import (
	"log"
	"testing"
)

func TestDatasetEvaluator(t *testing.T) {
	params := map[string]string{"script": ML_SCRIPT, "testset": TESTSET}
	eval, err := NewDatasetEvaluator(ONLINE_EVAL, params)
	if err != nil || eval == nil {
		log.Println(err)
		t.FailNow()
	}
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
