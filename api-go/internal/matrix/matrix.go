// Package matrix implementa la capa de Dominio (Domain Layer) de álgebra lineal:
// entidad de error de dominio, validación dimensional, rotación 90° horario y factorización QR.
package matrix

import (
	"fmt"
	"math"
)

// MaxDimension define la cota máxima de filas o columnas permitida por matriz
// para prevenir denegación de servicio (DoS) por agotamiento de memoria/CPU en LAPACK.
const MaxDimension = 250

// ValidationError es un error de dominio por entrada inválida. El adaptador HTTP lo traduce a 400.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// Validate exige una matriz rectangular no vacía de números finitos dentro de MaxDimension.
func Validate(m [][]float64) error {
	if len(m) == 0 {
		return &ValidationError{Message: "la matriz no puede estar vacía"}
	}
	cols := len(m[0])
	if cols == 0 {
		return &ValidationError{Message: "la matriz no puede tener filas vacías"}
	}
	if len(m) > MaxDimension || cols > MaxDimension {
		return &ValidationError{Message: fmt.Sprintf("la matriz excede el tamaño máximo permitido (%dx%d)", MaxDimension, MaxDimension)}
	}
	for i, row := range m {
		if len(row) != cols {
			return &ValidationError{Message: fmt.Sprintf(
				"la matriz no es rectangular: la fila %d tiene %d columnas y la primera tiene %d",
				i, len(row), cols,
			)}
		}
		for j, value := range row {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return &ValidationError{Message: fmt.Sprintf("el valor en [%d][%d] no es un número finito", i, j)}
			}
		}
	}
	return nil
}

// Rotate90Clockwise gira la matriz 90 grados en sentido horario.
// Una matriz m×n pasa a ser n×m. Los valores no cambian, solo su posición.
func Rotate90Clockwise(m [][]float64) [][]float64 {
	rows := len(m)
	cols := len(m[0])
	rotated := make([][]float64, cols)
	for j := 0; j < cols; j++ {
		rotated[j] = make([]float64, rows)
		for i := 0; i < rows; i++ {
			rotated[j][rows-1-i] = m[i][j]
		}
	}
	return rotated
}
