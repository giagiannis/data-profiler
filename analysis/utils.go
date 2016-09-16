package analysis

import "fmt"

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
