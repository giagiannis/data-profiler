package core

import (
	"math"
	"strings"
)

type DistanceType uint8

const (
	// represents the Euclidean distance
	EUCLIDEAN DistanceType = iota
	// represents the Manhattan distance
	MANHATTAN
	// represents the inverse expression of cosine similarity
	COSINE
	// represents the unknown distance
	UNKNOWN
)

func euclidean(x1, x2 []float64) float64 {
	sum := 0.0
	for i, _ := range x1 {
		sum += (x1[i] - x2[i]) * (x1[i] - x2[i])
	}
	return math.Sqrt(sum)
}

func manhattan(x1, x2 []float64) float64 {
	sum := 0.0
	for i, _ := range x1 {
		sum += math.Abs(x1[i] - x2[i])
	}
	return sum
}

func cosine(x1, x2 []float64) float64 {
	similarity := 0.0
	for i, _ := range x1 {
		similarity += x1[i] * x2[i]
	}
	norm := euclideanNorm(x1)
	if norm == 0 {
		return math.NaN()
	}
	similarity /= norm
	norm = euclideanNorm(x2)
	if norm == 0 {
		return math.NaN()
	}
	similarity /= euclideanNorm(x2)
	return math.Acos(similarity) / math.Pi
}

func euclideanNorm(x1 []float64) float64 {
	return euclidean(x1, make([]float64, len(x1)))
}

// Distance function is used to provide the distance between two different
// vectors. NaN is returned in the following cases: if the vectors are not
// comparable (e.g., have different number of dimensions), if the DistanceType
// is unknown, or if the norm of the vector is null.
func Distance(x1, x2 []float64, d DistanceType) float64 {
	if d == EUCLIDEAN {
		return euclidean(x1, x2)
	} else if d == MANHATTAN {
		return manhattan(x1, x2)
	} else if d == COSINE {
		return cosine(x1, x2)
	} else {
		return math.NaN()
	}
}

func DistanceParsers(dType string) DistanceType {
	if strings.ToLower(dType) == "euclidean" {
		return EUCLIDEAN
	} else if strings.ToLower(dType) == "manhattan" {
		return MANHATTAN
	} else if strings.ToLower(dType) == "cosine" {
		return COSINE
	}
	return UNKNOWN
}
