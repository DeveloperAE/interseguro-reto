package matrix

import (
	"math"

	"gonum.org/v1/gonum/blas/blas64"
	"gonum.org/v1/gonum/lapack/lapack64"
	"gonum.org/v1/gonum/mat"
)

// QR es la factorización A ≈ Q·R.
// Q es m×k con columnas ortonormales y R es k×n trapezoidal superior,
// con k = min(m, n). Sirve tanto si la matriz es alta como si es ancha.
type QR struct {
	Q                   [][]float64
	R                   [][]float64
	ReconstructionError float64
}

// FactorizeQR calcula la factorización QR reducida con LAPACK (gonum).
// La matriz de entrada no se modifica.
func FactorizeQR(values [][]float64) (QR, error) {
	if err := Validate(values); err != nil {
		return QR{}, err
	}
	rows := len(values)
	cols := len(values[0])
	original := mat.NewDense(rows, cols, nil)
	factor := mat.NewDense(rows, cols, nil)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			original.Set(i, j, values[i][j])
			factor.Set(i, j, values[i][j])
		}
	}

	k := rows
	if cols < k {
		k = cols
	}
	tau := make([]float64, k)
	work := lapackWorkspace(factor.RawMatrix(), tau, true)
	lapack64.Geqrf(factor.RawMatrix(), tau, work, len(work))

	r := mat.NewDense(k, cols, nil)
	for i := 0; i < k; i++ {
		for j := i; j < cols; j++ {
			r.Set(i, j, factor.At(i, j))
		}
	}

	// Orgqr exige que el número de columnas no supere al de filas.
	// Por eso Q se arma solo con las k columnas de los reflectores.
	q := mat.NewDense(rows, k, nil)
	for i := 0; i < rows; i++ {
		for j := 0; j < k; j++ {
			q.Set(i, j, factor.At(i, j))
		}
	}
	qWork := lapackWorkspace(q.RawMatrix(), tau, false)
	lapack64.Orgqr(q.RawMatrix(), tau, qWork, len(qWork))

	var product mat.Dense
	product.Mul(q, r)
	var diff mat.Dense
	diff.Sub(original, &product)

	return QR{
		Q:                   denseToSlices(q),
		R:                   denseToSlices(r),
		ReconstructionError: frobeniusNorm(&diff),
	}, nil
}

// frobeniusNorm calcula la norma de Frobenius ||A||_F = sqrt(sum(a_ij^2)) en O(m*n),
// evitando el costo cúbico de calcular valores singulares (SVD) con mat.Norm(d, 2).
func frobeniusNorm(d *mat.Dense) float64 {
	raw := d.RawMatrix()
	var sumSq float64
	for i := 0; i < raw.Rows; i++ {
		row := raw.Data[i*raw.Stride : i*raw.Stride+raw.Cols]
		for _, v := range row {
			sumSq += v * v
		}
	}
	return math.Sqrt(sumSq)
}

func lapackWorkspace(a blas64.General, tau []float64, geqrf bool) []float64 {
	probe := []float64{0}
	if geqrf {
		lapack64.Geqrf(a, tau, probe, -1)
	} else {
		lapack64.Orgqr(a, tau, probe, -1)
	}
	size := int(math.Ceil(probe[0]))
	if size < 1 {
		size = 1
	}
	return make([]float64, size)
}

func denseToSlices(d *mat.Dense) [][]float64 {
	rows, cols := d.Dims()
	out := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		out[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			out[i][j] = d.At(i, j)
		}
	}
	return out
}
