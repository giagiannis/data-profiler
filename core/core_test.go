package core

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	TRAINSET        = "../_testdata/datatraining.csv"
	TESTSET         = "../_testdata/datatest.csv"
	ANALYSIS_SCRIPT = "../_rscripts/pca.R"
	ML_SCRIPT       = "../_rscripts/lm.R"
	TMP_DIR         = "/tmp/"
)

func init() {
	log.SetOutput(ioutil.Discard)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	rand.Seed(int64(time.Now().Nanosecond()))
}

// Creates a random tuple with floats, containing fields fields.
func randomTupleGenerator(fields int) []float64 {
	res := make([]float64, fields)
	for i := 0; i < fields; i++ {
		res[i] = rand.Float64()
	}
	return res
}

// Creates a pool of tuples (its sizeis determined by poolSize) and a number of
// datasets each containing tuples with the specified number of attributes.
func createPoolBasedDatasets(poolSize, datasets, attributes int) []*Dataset {
	pool := *new([][]float64)
	for i := 0; i < poolSize; i++ {
		pool = append(pool, randomTupleGenerator(attributes))
	}
	fileNames := make([]string, datasets)
	for i := 0; i < datasets; i++ {
		dSize := rand.Int()%(poolSize/2) + poolSize/2
		buffer := make([]byte, 4)
		rand.Read(buffer)
		fileNames[i] = fmt.Sprintf("%s%x.txt", TMP_DIR, buffer)
		f, _ := os.Create(fileNames[i])
		defer f.Close()

		builder := bytes.Buffer{}
		for i := 0; i < attributes-1; i++ {
			builder.WriteString(fmt.Sprintf("x%d, ", i))
		}
		builder.WriteString("class\n")
		tuplesChosen := make(map[int]bool)
		for j := 0; j < dSize; j++ {
			var idx int
			for idx = rand.Int() % len(pool); tuplesChosen[idx]; idx = rand.Int() % len(pool) {
			}
			tuplesChosen[idx] = true
			tuple := pool[idx]
			tupleLength := len(tuple)
			for i, v := range tuple {
				builder.WriteString(fmt.Sprintf("%.5f", v))
				if i < tupleLength-1 {
					builder.WriteString(",")
				}
			}
			builder.WriteString("\n")
		}
		f.Write(builder.Bytes())
	}
	result := make([]*Dataset, len(fileNames))
	for i, f := range fileNames {
		result[i] = NewDataset(f)
	}

	return result
}
