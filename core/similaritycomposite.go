package core

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
)

// CompositeEstimator returns a similarity function based on
// compositions of simpler similarity expressions. The user needs to provide a
// formula containing the expression of the similarity function, e.g.:
// 0.8 x BHATTACHARRYA + 0.2 CORRELATION
// Note that it is the user's responsibility to guarantee that the overall
// expression remains within the limits of the similarity expression [0,1].
type CompositeEstimator struct {
	AbstractDatasetSimilarityEstimator

	// the slice of estimators to use for the similarity evaluation
	estimators map[string]DatasetSimilarityEstimator
	// the math expression used for the estimation
	expression string
}

// Compute method constructs the Similarity Matrix
func (e *CompositeEstimator) Compute() error {
	for _, est := range e.estimators {
		est.Compute()
	}
	return datasetSimilarityEstimatorCompute(e)
}

// Similarity returns the similarity between two datasets
func (e *CompositeEstimator) Similarity(a, b *Dataset) float64 {
	expression, err := govaluate.NewEvaluableExpression(e.expression)
	if err != nil {
		log.Println(err)
		return 0.0
	}
	params := make(map[string]interface{})
	for k, est := range e.estimators {
		params[k] = est.Similarity(a, b)
	}
	result, err := expression.Evaluate(params)
	if err != nil {
		log.Println(err)
		return 0.0
	}
	if val, ok := result.(float64); ok {
		return val
	}
	return -1.0
}

// Serialize returns an array of bytes representing the Estimator.
func (e *CompositeEstimator) Serialize() []byte {
	buffer := new(bytes.Buffer)
	buffer.Write(getBytesInt(int(SimilarityTypeComposite)))

	buffer.Write(datasetSimilarityEstimatorSerialize(
		e.AbstractDatasetSimilarityEstimator))

	// serialize expression
	buffer.WriteString(e.expression + "\n")

	// serialize estimators
	buffer.Write(getBytesInt(len(e.estimators)))
	for k, est := range e.estimators {
		buffer.WriteString(k + "\n")
		temp := est.Serialize()
		buffer.Write(getBytesInt(len(temp)))
		buffer.Write(temp)
	}

	return buffer.Bytes()
}

// Deserialize constructs an Estimator object based on the byte array provided.
func (e *CompositeEstimator) Deserialize(b []byte) {
	buffer := bytes.NewBuffer(b)
	tempInt := make([]byte, 4)
	buffer.Read(tempInt) // consume estimator type
	buffer.Read(tempInt)
	absEstBytes := make([]byte, getIntBytes(tempInt))
	buffer.Read(absEstBytes)
	e.AbstractDatasetSimilarityEstimator =
		*datasetSimilarityEstimatorDeserialize(absEstBytes)

	// parse expression
	line, _ := buffer.ReadString('\n')
	e.expression = strings.TrimSpace(line)

	// parse the estimators
	buffer.Read(tempInt)
	count := getIntBytes(tempInt)
	e.estimators = make(map[string]DatasetSimilarityEstimator)
	for i := 0; i < count; i++ {
		line, _ := buffer.ReadString('\n')
		key := strings.TrimSpace(line)
		buffer.Read(tempInt)
		length := getIntBytes(tempInt)
		tempBuff := make([]byte, length)
		buffer.Read(tempBuff)
		est := DeserializeSimilarityEstimator(tempBuff)
		e.estimators[key] = est
	}
}

// Configure provides the configuration parameters needed by the Estimator
func (e *CompositeEstimator) Configure(conf map[string]string) {
	e.estimators = make(map[string]DatasetSimilarityEstimator)
	for k, v := range conf {
		if k == "concurrency" {
			val, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				log.Println(err)
			}
			e.concurrency = int(val)
		} else if k == "expression" {
			log.Println(v)
			e.expression = v
		} else {
			estConf := DeserializeConfigurationOptions(v, "|")
			if val, ok := estConf["type"]; ok {
				estType := NewDatasetSimilarityEstimatorType(val)
				if estType != nil {
					est := NewDatasetSimilarityEstimator(*estType, e.datasets)
					est.Configure(estConf)
					e.estimators[k] = est
				} else {
					log.Println("Provided operator does not exist!")
				}
			} else {
				log.Println("Cannot initialize estiamator witout type")
			}
		}
	}
}

// Options returns the applicable parameters needed by the Estimator.
func (e *CompositeEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency": "max number of threads to be run in parallel",
		"expression": "the math expression that combines the estimators " +
			"e.g.: 0.2*x + 0.8*y " +
			"(x and y must be later defined)",
		"x": "the conf parameters for x, separated by | e.g." +
			"type:bhattacharyya|concurrency:10",
		"y": "the conf parameters for y, separated by | e.g." +
			"type:correlation|type:pearson",
	}
}

// SerializeConfigurationOptions is used to transform a map holding the conf
// options into a map
func SerializeConfigurationOptions(conf map[string]string, sep string) string {
	output := ""
	i := 0
	for k, v := range conf {
		output += fmt.Sprintf("%s:%s", k, v)
		if i < len(conf)-1 {
			output += sep
		}
		i++
	}
	return output
}

// DeserializeConfigurationOptions generates a map holding configuration options
// based on the serialized form.
func DeserializeConfigurationOptions(serialized, sep string) map[string]string {
	result := make(map[string]string)
	arr := strings.Split(serialized, sep)
	for i := range arr {
		temp := strings.Split(arr[i], ":")
		if len(temp) != 2 {
			return nil
		}
		result[temp[0]] = temp[1]
	}
	return result
}
