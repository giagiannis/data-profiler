package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
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
	if t == nil {
		return
	}
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
	Dataset     *ModelDataset
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
	task.Dataset = dts
	task.Description = fmt.Sprintf("SM Computation for %s, type %s\n",
		dts.Name, conf["estimatorType"])
	task.fnc = func() error {
		datasets := core.DiscoverDatasets(dts.Path)
		estType := *core.NewDatasetSimilarityEstimatorType(conf["estimatorType"])
		est := core.NewDatasetSimilarityEstimator(estType, datasets)
		//TODO: take into consideration the population policy
		est.Configure(conf)
		err := est.Compute()
		if err != nil {
			return err
		}
		sm := est.SimilarityMatrix()
		//		var smID string
		if sm != nil {
			//smID =
			modelSimilarityMatrixInsert(datasetID, sm.Serialize(), conf)
		}
		//modelEstimatorInsert(datasetID, smID, est.Serialize(), conf)
		return nil
	}
	return task
}

func NewMDSComputationTask(smID, datasetID string, conf map[string]string) *Task {
	smModel := modelSimilarityMatrixGet(smID)
	if smModel == nil {
		log.Println("SM not found")
		return nil
	}

	cnt, err := ioutil.ReadFile(smModel.Path)
	if err != nil {
		log.Println(err)
	}
	sm := new(core.DatasetSimilarityMatrix)
	sm.Deserialize(cnt)
	k, err := strconv.ParseInt(conf["k"], 10, 64)
	if err != nil {
		log.Println(err)
	}

	dat := modelDatasetGetInfo(datasetID)
	task := new(Task)
	task.Dataset = dat
	task.Description = fmt.Sprintf("MDS Execution for %s with k=%d\n",
		dat.Name, k)
	task.fnc = func() error {
		mds := core.NewMDScaling(sm, int(k), Conf.Scripts.MDS)
		err = mds.Compute()
		if err != nil {
			return err
		}
		// TODO: write this to the database
		gof := fmt.Sprintf("%.5f", mds.Gof())
		modelCoordinatesInsert(mds.Coordinates(), dat.ID, conf["k"], gof, smID)
		return nil
	}
	return task
}
