package core

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestCompositeCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(10000, 10, 3)
	var configuration = map[string]string{
		"concurrency": "10",
		"expression":  "0.8*x + 0.2*y",
		"x":           "type:bhattacharyya|tree.scale:0.5",
		"y":           "type:jaccard|concurrency:10",
	}
	e := new(CompositeEstimator)
	e.datasets = datasets
	e.Configure(configuration)

	err := e.Compute()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	sm := e.SimilarityMatrix()
	smSanityCheck(sm, t)
	cleanDatasets(datasets)
}

func TestCompositeConfiguration(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 100, 3)
	e := createCompositeEstimator(datasets)
	if len(e.estimators) != 2 {
		t.Logf("Expected %d estimators and found %d\n", 2, len(e.estimators))
		t.Fail()
	}
	if _, ok := (e.estimators["x"]).(*BhattacharyyaEstimator); !ok {
		t.Logf("Expected type %s was not identified\n",
			"*core.BhattacharyyaEstimator")
		t.Fail()
	}
	if _, ok := (e.estimators["y"]).(*JaccardEstimator); !ok {
		t.Logf("Expected type %s was not identified\n",
			"*core.JaccardEstimator")
		t.Fail()
	}

	cleanDatasets(datasets)
}

func TestCompositeSerialization(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 10, 3)
	e := createCompositeEstimator(datasets)
	e.Compute()
	ser := e.Serialize()
	if ser == nil {
		t.Log("Nil serialization object")
		t.Fail()
	}
	newEst := new(CompositeEstimator)
	newEst.Deserialize(ser)
	if newEst.expression != e.expression {
		t.Log("Difference in expressions")
		t.Fail()
	}
	for k, est := range newEst.estimators {
		if val, ok := e.estimators[k]; ok {
			if reflect.TypeOf(val) != reflect.TypeOf(est) {
				t.Log("Estimator types are different")
				t.Fail()
			}
		} else {
			t.Log("Estimator not found")
			t.Fail()
		}
	}
	estimatorsCheck(newEst.AbstractDatasetSimilarityEstimator,
		e.AbstractDatasetSimilarityEstimator, t)
	cleanDatasets(datasets)
}

func TestConfigurationOptionsSerialization(t *testing.T) {
	letters := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k",
		"l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	randomString := func(count int) string {
		res := ""
		for i := 0; i < count; i++ {
			res += letters[rand.Int()%len(letters)]
		}
		return res
	}
	sep := "|"
	conf := make(map[string]string)
	numOfOptions := rand.Int() % 30
	for i := 0; i < numOfOptions; i++ {
		key, value := randomString(10), randomString(10)
		conf[key] = value
	}
	output := SerializeConfigurationOptions(conf, sep)
	newConf := DeserializeConfigurationOptions(output, sep)
	if len(conf) != len(newConf) {
		t.Logf("Expected conf of size %d, found %d\n", len(conf), len(newConf))
		t.Fail()
	}
	for k, v := range conf {
		if v != newConf[k] {
			t.Logf("Expected value %s, found %s\n", v, newConf[k])
			t.Fail()
		}
	}
}

func createCompositeEstimator(datasets []*Dataset) *CompositeEstimator {
	var configuration = map[string]string{
		"concurrency": "10",
		"expression":  "0.8*x + 0.2*y",
		"x":           "type:bhattacharyya|tree.scale:0.5",
		"y":           "type:jaccard|concurrency:10",
	}
	e := new(CompositeEstimator)
	e.datasets = datasets
	e.Configure(configuration)
	return e
}
