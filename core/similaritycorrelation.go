package core

import (
	"bytes"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

// CorrelationEstimator estimates the similarity between two datasets based
// on a correlation metric. This metric can only be used for datasets that
// consist of a single column and consist of the same number of tuples.
type CorrelationEstimator struct {
	AbstractDatasetSimilarityEstimator
	estType CorrelationEstimatorType
}

// CorrelationEstimatorType represents the type of correlation to be used
// by the CorrelationEstimator
type CorrelationEstimatorType uint8

const (
	// CorrelationSimilarityTypePearson represents the Pearson cor. coeff
	CorrelationSimilarityTypePearson = iota
	// CorrelationSimilarityTypeSpearman represents the Spearman cor. coeff
	CorrelationSimilarityTypeSpearman = iota + 1
	// CorrelationSimilarityTypeKendall represents the Kendall cor. coeff
	CorrelationSimilarityTypeKendall = iota + 2
)

// Configure provides a set of configuration options to the CorrelationEstimator
// struct.
func (e *CorrelationEstimator) Configure(conf map[string]string) {
	if val, ok := conf["concurrency"]; ok {
		conv, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			log.Println(err)
		}
		e.concurrency = int(conv)
	}
	if val, ok := conf["type"]; ok {
		if strings.ToLower(val) == "pearson" {
			e.estType = CorrelationSimilarityTypePearson
		} else if strings.ToLower(val) == "spearman" {
			e.estType = CorrelationSimilarityTypeSpearman
		} else if strings.ToLower(val) == "kendall" {
			e.estType = CorrelationSimilarityTypeKendall
		}
	}
}

// Options returns a list of options used internally by the CorrelationEstimator
// struct for its execution.
func (e *CorrelationEstimator) Options() map[string]string {
	return map[string]string{
		"concurrency": "max number of threads to use",
		"type":        "one of [Pearson], Spearman, Kendall",
	}
}

// Serialize returns a byte array in order to serialize the CorrelationEstimator
// struct.
func (e *CorrelationEstimator) Serialize() []byte {
	buffer := new(bytes.Buffer)
	buffer.Write(getBytesInt(int(SimilarityTypeCorrelation)))
	buffer.Write(
		datasetSimilarityEstimatorSerialize(
			e.AbstractDatasetSimilarityEstimator))
	buffer.Write(getBytesInt(int(e.estType)))
	return buffer.Bytes()
}

// Deserialize returns a byte array in order to deserialize the CorrelationEstimator
func (e *CorrelationEstimator) Deserialize(b []byte) {
	buffer := bytes.NewBuffer(b)
	tempInt := make([]byte, 4)
	buffer.Read(tempInt) // consume estimator type
	buffer.Read(tempInt)
	absEstBytes := make([]byte, getIntBytes(tempInt))
	buffer.Read(absEstBytes)
	e.AbstractDatasetSimilarityEstimator =
		*datasetSimilarityEstimatorDeserialize(absEstBytes)

	buffer.Read(tempInt)
	e.estType = CorrelationEstimatorType(getIntBytes(tempInt))
}

// Compute runs the CorrelationEstimator for the provided datasets and generates
// the SimilarityMatrix object
func (e *CorrelationEstimator) Compute() error {
	e.similarities = NewDatasetSimilarities(len(e.datasets))

	log.Println("Fetching datasets in memory")
	if e.datasets == nil || len(e.datasets) == 0 {
		log.Println("No datasets were given")
		return errors.New("Empty dataset slice")
	}
	for _, d := range e.datasets {
		err := d.ReadFromFile()
		if err != nil {
			log.Println(err)
			return err
		}
	}

	start := time.Now()
	datasetSimilarityEstimatorCompute(e)
	e.duration = time.Since(start).Seconds()
	return nil
}

// Similarity returns the similarity between two datasets. Since all the
// correlation coefficients are between [-1.0,1.0], the output of this function
// is scaled to [0.0,1.0] by returning (x/2.0 + 0.5), where x is one of
// Pearson, Spearman and Kendall coefficients.
func (e *CorrelationEstimator) Similarity(a, b *Dataset) float64 {
	aTrans, bTrans := e.transformDataset(a), e.transformDataset(b)
	if e.estType == CorrelationSimilarityTypePearson {
		return Pearson(aTrans, bTrans)/2.0 + 0.5
	} else if e.estType == CorrelationSimilarityTypeSpearman {
		return Spearman(aTrans, bTrans)/2.0 + 0.5
	} else if e.estType == CorrelationSimilarityTypeKendall {
		return Kendall(aTrans, bTrans)/2.0 + 0.5
	}
	return .0
}

func (e *CorrelationEstimator) transformDataset(d *Dataset) []float64 {
	var result []float64
	for _, t := range d.Data() {
		result = append(result, t.Data[0])
	}
	return result
}

// Pearson returns the pearson correlation coefficient between two variables.
// The two arrays must be of the same size, else 0 is returned.
func Pearson(a, b []float64) float64 {
	aMean, bMean := Mean(a), Mean(b)
	if len(a) != len(b) {
		return 0.0
	}
	nom, denomA, denomB := .0, .0, .0
	for i := range a {
		nom += (a[i] - aMean) * (b[i] - bMean)
		denomA += (a[i] - aMean) * (a[i] - aMean)
		denomB += (b[i] - bMean) * (b[i] - bMean)
	}
	if (denomA * denomB) != 0 {
		return nom / (denomA * denomB)
	}
	return .0
}

// Spearman returns the rank correlation coefficient between two variables.
// The two arrays must be of the same size, else 0 is returned.
func Spearman(a, b []float64) float64 {
	aRanks, bRanks := Rank(a), Rank(b)
	if len(a) != len(b) {
		return 0.0
	}
	// stupid casting is required to use Pearson
	aRanksFloat := make([]float64, len(aRanks))
	bRanksFloat := make([]float64, len(bRanks))
	for i := range aRanks {
		aRanksFloat[i] = float64(aRanks[i])
		bRanksFloat[i] = float64(bRanks[i])
	}
	return Pearson(aRanksFloat, bRanksFloat)
}

// Kendall returns yet another rank correlation coefficient between two
// variables. The two arrays must be of the same size, else 0 is returned.
func Kendall(a, b []float64) float64 {
	aRanks, bRanks := Rank(a), Rank(b)
	if len(a) != len(b) {
		return 0.0
	}
	concordant, discordant := 0, 0
	size := len(aRanks)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if (aRanks[i]-aRanks[j])*(bRanks[i]-bRanks[j]) > 0 {
				concordant++
			} else if (aRanks[i]-aRanks[j])*(bRanks[i]-bRanks[j]) < 0 {
				discordant++
			} // else neither concordant nor discordant
		}
	}
	return float64(concordant-discordant) / float64(size*(size-1)/2.0)
}

// Mean returns the mean value of a float array
func Mean(a []float64) float64 {
	if len(a) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range a {
		sum += v
	}
	return sum / float64(len(a))
}

// Rank returns an array containing the ranks of a slice a
func Rank(a []float64) []int {
	rank := make([]int, len(a))
	for i := range a {
		for j := range a {
			if a[j] < a[i] {
				rank[i]++
			}
		}
	}
	return rank
}
