package core

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestCompositeCompute(t *testing.T) {
}

func TestCompositeConfiguration(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 100, 3)
	var configuration = map[string]string{
		"concurrency": "10",
		"expression":  "0.8*x + 0.2*y",
		"x":           "type=bhattacharyya|tree.scale=0.8",
		"y":           "type=jaccard|concurrency=10",
	}
	e := new(CompositeEstimator)
	e.datasets = datasets
	e.Configure(configuration)
	if len(e.estimators) != 2 {
		t.Logf("Expected %d estimators and found %d\n", 2, e.estimators)
		t.Fail()
	}
	if reflect.TypeOf(e.estimators["x"]).String() != "*core.BhattacharyyaEstimator" {
		t.Logf("Expected type %s and found %s\n",
			"*core.BhattacharyyaEstimator",
			reflect.TypeOf(e.estimators["x"]).String())
		t.Fail()
	}
	if reflect.TypeOf(e.estimators["y"]).String() != "*core.JaccardEstimator" {
		t.Logf("Expected type %s and found %s\n",
			"*core.JaccardEstimator",
			reflect.TypeOf(e.estimators["y"]).String())
		t.Fail()
	}

	cleanDatasets(datasets)
}

func TestDeSerializeConfigurationOptions(t *testing.T) {
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
