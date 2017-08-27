package core

import (
	"encoding/binary"
	"errors"
	"math"
)

func getBytesInt(val int) []byte {
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, uint32(val))
	return temp
}

func getBytesFloat(val float64) []byte {
	bits := math.Float64bits(val)
	temp := make([]byte, 8)
	binary.BigEndian.PutUint64(temp, bits)
	return temp
}

func getIntBytes(buf []byte) int {
	return int(binary.BigEndian.Uint32(buf))
}
func getFloatBytes(buf []byte) float64 {
	bits := binary.BigEndian.Uint64(buf)
	float := math.Float64frombits(bits)
	return float
}

func norm(a, b []float64, normDegree int) (float64, error) {
	if len(a) != len(b) {
		return -1, errors.New("arrays have different sizes")
	}
	sum := 0.0
	for i := range a {
		dif := math.Abs(a[i] - b[i])
		sum += math.Pow(dif, float64(normDegree))
	}
	return math.Pow(sum, 1.0/float64(normDegree)), nil
}
