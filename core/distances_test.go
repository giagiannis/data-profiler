package core

import (
	"math"
	"math/rand"
	"testing"
)

func TestEuclidean(t *testing.T) {
	dim := rand.Int() % 200
	x1 := make([]float64, dim)
	x2 := make([]float64, dim)
	for i := 0; i < dim; i++ {
		x1[i] = rand.Float64()
		x2[i] = rand.Float64()
	}
	eucl := Distance(x1, x2, EUCLIDEAN)
	manh := Distance(x1, x2, MANHATTAN)
	if eucl > manh {
		t.Log("Euclidean > Manhattan")
		t.FailNow()
	}
	cos := Distance(x1, x2, COSINE)
	if euclideanNorm(x1) > 0 && euclideanNorm(x2) > 0 && cos == math.NaN() {
		t.Log("Cosine distance is NaN but vector norms are positive")
		t.FailNow()
	}
}
