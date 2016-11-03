package core

import (
	"fmt"
	"math"
	"strings"
)

type DimensionEnergyCollection struct {
	dim    []int
	energy []float64
	acc    float64
	cutoff float64
}

func NewDimensionEnergyCollection(energies []float64, cutoff float64) *DimensionEnergyCollection {
	ret := new(DimensionEnergyCollection)
	ret.energy = energies
	ret.dim = make([]int, len(energies))
	ret.acc = 0.0
	for i := 0; i < len(ret.dim); i++ {
		ret.dim[i] = i
		ret.acc += ret.energy[i]
	}
	ret.cutoff = cutoff
	return ret
}

func (kn *DimensionEnergyCollection) Len() int {
	return len(kn.dim)
}

func (kn *DimensionEnergyCollection) Less(i, j int) bool {
	return kn.energy[i] < kn.energy[j]
}

func (kn *DimensionEnergyCollection) Swap(i, j int) {
	tempInd := kn.dim[i]
	tempEnergy := kn.energy[i]

	kn.dim[i] = kn.dim[j]
	kn.energy[i] = kn.energy[j]

	kn.dim[j] = tempInd
	kn.energy[j] = tempEnergy
}
func (kn *DimensionEnergyCollection) String() string {
	res := ""
	for i := range kn.dim {
		res += fmt.Sprintf("[%d -> %.5f]", kn.dim[i], kn.energy[i])
	}
	return res
}

// returns the set of indices that present an accumulative energy within the
// cutoff threshold
func (kn *DimensionEnergyCollection) Cutoff() []int {
	acc := 0.0
	results := make([]int, 0, len(kn.dim))
	for i, v := range kn.energy {
		if acc <= kn.cutoff*kn.acc {
			results = append(results, kn.dim[i])
			acc += v
		}
	}
	return results
}

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
