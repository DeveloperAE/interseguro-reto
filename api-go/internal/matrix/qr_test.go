package matrix

import (
	"math"
	"testing"
)

func TestFactorizeQR(t *testing.T) {
	matrices := [][][]float64{
		{{12, -51}, {6, 167}, {-4, 24}},
		{{1, 2, 3}, {4, 5, 6}},
		{{1, 2}, {3, 4}, {5, 6}},
		{{1, 0}, {0, 2}},
		{{5}},
		{{1, 2, 3}},
	}
	for _, values := range matrices {
		got, err := FactorizeQR(values)
		if err != nil {
			t.Fatal(err)
		}
		if got.ReconstructionError > 1e-8 {
			t.Fatalf("error de reconstrucción %g demasiado alto para %v", got.ReconstructionError, values)
		}
		if err := orthonormalColumns(got.Q); err != nil {
			t.Fatalf("Q no tiene columnas ortonormales para %v: %v", values, err)
		}
	}

	if _, err := FactorizeQR(nil); err == nil {
		t.Fatal("se esperaba error al factorizar matriz vacía")
	}
}

func orthonormalColumns(q [][]float64) error {
	rows := len(q)
	cols := len(q[0])
	for a := 0; a < cols; a++ {
		for b := a; b < cols; b++ {
			dot := 0.0
			for i := 0; i < rows; i++ {
				dot += q[i][a] * q[i][b]
			}
			want := 0.0
			if a == b {
				want = 1
			}
			if math.Abs(dot-want) > 1e-8 {
				return errDot
			}
		}
	}
	return nil
}

var errDot = dotError{}

type dotError struct{}

func (dotError) Error() string { return "producto interno fuera de tolerancia" }
