package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/giagiannis/data-profiler/core"
)

// TaskEngine is deployed once for the server's lifetime and keeps the tasks
// which are executed.
type TaskEngine struct {
	Tasks []*Task
	lock  sync.Mutex
}

func NewTaskEngine() *TaskEngine {
	te := new(TaskEngine)
	te.lock = *new(sync.Mutex)
	return te
}

func (e *TaskEngine) Submit(t *Task) {
	e.lock.Lock()
	go t.Run()
	e.Tasks = append(e.Tasks, t)
	e.lock.Unlock()
}

// Task is the primitive struct that represents a task of the server.
type Task struct {
	Status      string
	Started     time.Time
	Duration    float64
	Description string
	fnc         func() error
}

func (t *Task) Run() {
	t.Started = time.Now()
	t.Status = "RUNNING"
	err := t.fnc()
	if err != nil {
		t.Status = "ERROR - " + err.Error()
	} else {
		t.Status = "DONE"
		t.Duration = time.Since(t.Started).Seconds()
	}
}

func NewSMComputationTask(datasetID string, conf map[string]string) *Task {
	task := new(Task)
	dts := modelDatasetGetInfo(datasetID)
	task.Description = fmt.Sprintf("SM Computation for %s, type %s\n",
		dts.Name, conf["type"])
	task.fnc = func() error {
		datasets := core.DiscoverDatasets(dts.Path)
		estType := *core.NewDatasetSimilarityEstimatorType(conf["type"])
		est := core.NewDatasetSimilarityEstimator(estType, datasets)
		est.Configure(conf)
		err := est.Compute()
		if err != nil {
			return err
		}
		// TODO: serialize them and write them to the DB
		//est.Serialize()
		//est.SimilarityMatrix().Serialize()
		return nil
	}
	return task
}
