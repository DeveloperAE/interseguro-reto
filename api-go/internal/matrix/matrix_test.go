package matrix

import (
	"math"
	"testing"
)

func TestValidate(t *testing.T) {
	oversized := make([][]float64, MaxDimension+1)
	for i := range oversized {
		oversized[i] = []float64{1}
	}
	tests := []struct {
		name    string
		matrix  [][]float64
		wantErr bool
	}{
		{name: "rectangular", matrix: [][]float64{{1, 2, 3}, {4, 5, 6}}},
		{name: "vacia", matrix: nil, wantErr: true},
		{name: "fila vacia", matrix: [][]float64{{}}, wantErr: true},
		{name: "irregular", matrix: [][]float64{{1, 2}, {3}}, wantErr: true},
		{name: "nan", matrix: [][]float64{{math.NaN()}}, wantErr: true},
		{name: "infinito", matrix: [][]float64{{math.Inf(1)}}, wantErr: true},
		{name: "excede dimension maxima", matrix: oversized, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.matrix)
			if tt.wantErr && err == nil {
				t.Fatal("se esperaba un error")
			}
			if tt.wantErr && err.Error() == "" {
				t.Fatal("mensaje de error vacío")
			}
			if !tt.wantErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRotate90Clockwise(t *testing.T) {
	tests := []struct {
		name string
		in   [][]float64
		want [][]float64
	}{
		{
			name: "2x3",
			in:   [][]float64{{1, 2, 3}, {4, 5, 6}},
			want: [][]float64{{4, 1}, {5, 2}, {6, 3}},
		},
		{
			name: "2x2",
			in:   [][]float64{{1, 2}, {3, 4}},
			want: [][]float64{{3, 1}, {4, 2}},
		},
		{
			name: "1x3",
			in:   [][]float64{{1, 2, 3}},
			want: [][]float64{{1}, {2}, {3}},
		},
		{
			name: "3x1",
			in:   [][]float64{{1}, {2}, {3}},
			want: [][]float64{{3, 2, 1}},
		},
		{
			// Una diagonal, al rotarla 90°, deja de ser diagonal.
			name: "diagonal deja de serlo",
			in:   [][]float64{{1, 0}, {0, 2}},
			want: [][]float64{{0, 1}, {2, 0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Rotate90Clockwise(tt.in)
			if !equalMatrix(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func equalMatrix(a, b [][]float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
