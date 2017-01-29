package core

import (
	"encoding/binary"
	"math"
)

// Returns the similarity based on the distance
func DistanceToSimilarity(distance float64) float64 {
	return 1.0 / (1.0 + distance)
}

// Returns the distance based on the similarity
func SimilarityToDistance(similarity float64) float64 {
	return 1.0/similarity - 1.0
}

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
