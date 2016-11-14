package core

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type DatasetEvaluatorType uint8

const (
	ONLINE_EVAL DatasetEvaluatorType = iota + 1
	FILE_BASED_EVAL
)

type DatasetEvaluator interface {
	Evaluate(string) (float64, error)
}

// Returns a new DatasetEvaluator object
func NewDatasetEvaluator(evalType DatasetEvaluatorType,
	params map[string]string) (DatasetEvaluator, error) {
	if evalType == ONLINE_EVAL {
		eval := new(OnlineDatasetEvaluator)
		if _, ok := params["script"]; !ok {
			return nil, errors.New("Online evaluator needs script param")
		}
		eval.script = params["script"]
		if _, ok := params["testset"]; !ok {
			return nil, errors.New("Online evaluator needs testset param")
		}
		eval.testset = params["testset"]
		return eval, nil
	} else if evalType == FILE_BASED_EVAL {
		eval := new(FileBasedEvaluator)
		if _, ok := params["scores"]; !ok {
			return nil, errors.New("File based evaluator needs scores file")
		}
		f, err := os.Open(params["scores"])
		if err != nil {
			return nil, err
		}
		scores := NewDatasetScores()
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		err = scores.Deserialize(buf)
		if err != nil {
			return nil, err
		}
		eval.scores = *scores
		return eval, nil
	}
	return nil, errors.New("Not suitable evaluator found")
}

// DatasetEvaluator is responsible to execute the training script and fetch
// the model accuracy
type OnlineDatasetEvaluator struct {
	script  string
	testset string
}

// Trains and evaluates a new dataset, based on its path. Returns an error
// estimation.
func (e *OnlineDatasetEvaluator) Evaluate(dataset string) (float64, error) {
	cmd := exec.Command(e.script, dataset, e.testset)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		return -1, err
	}
	val, err := strconv.ParseFloat(string(out), 64)
	if err != nil {
		log.Println(err)
		return -1, err
	}
	return val, nil
}

type FileBasedEvaluator struct {
	scores DatasetScores
}

func (e *FileBasedEvaluator) Evaluate(dataset string) (float64, error) {
	val, ok := e.scores.Scores[dataset]
	if !ok {
		return -1, errors.New("Dataset not found")
	}
	return val, nil
}

// DatasetScores is used to store the scores of a set of datasets
type DatasetScores struct {
	Scores map[string]float64
}

func NewDatasetScores() *DatasetScores {
	o := new(DatasetScores)
	o.Scores = make(map[string]float64)
	return o
}

func (s *DatasetScores) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	e := gob.NewEncoder(buf)
	err := e.Encode(s.Scores)
	if err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}

func (s *DatasetScores) Deserialize(buf []byte) error {
	content := bytes.NewBuffer(buf)
	d := gob.NewDecoder(content)
	err := d.Decode(s.Scores)
	return err
}
