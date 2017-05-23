package core

import (
	"log"
	"testing"
)

func TestDatasetEvaluator(t *testing.T) {
	params := map[string]string{"script": mlScript, "testset": testSet}
	eval, err := NewDatasetEvaluator(OnlineEval, params)
	if err != nil || eval == nil {
		log.Println(err)
		t.FailNow()
	}
	val, err := eval.Evaluate(trainSet)
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if val < 0 {
		log.Println("Wrong conversion")
		t.FailNow()
	}
}
