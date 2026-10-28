// Package usecase contiene la capa de Aplicación (Casos de Uso y Puertos)
// siguiendo Arquitectura Hexagonal (Ports & Adapters).
package usecase

import (
	"context"
	"fmt"

	"api-go/internal/matrix"
	"api-go/internal/stats"
)

// StatisticsPort es el puerto de salida (Outbound Port) hacia el servicio de estadísticas.
type StatisticsPort interface {
	Statistics(ctx context.Context, authorization string, matrices []stats.Matrix) (stats.Result, error)
}

// UpstreamError envuelve fallos del servicio aguas abajo (puerto de salida)
// para que el adaptador primario HTTP devuelva 502 Bad Gateway sin acoplarse a detalles de red.
type UpstreamError struct {
	Err error
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("servicio de estadísticas no disponible: %v", e.Err)
}

func (e *UpstreamError) Unwrap() error {
	return e.Err
}

// Output representa el resultado de dominio y aplicación del procesamiento de una matriz.
type Output struct {
	Original   [][]float64
	Rotated    [][]float64
	Rotation   string
	QR         matrix.QR
	Statistics stats.Result
}

// ProcessMatrixUseCase orquesta la validación, rotación 90°, factorización QR
// y consulta de estadísticas a través del puerto de salida StatisticsPort.
type ProcessMatrixUseCase struct {
	statistics StatisticsPort
}

// NewProcessMatrix construye el caso de uso inyectando el puerto de estadísticas.
func NewProcessMatrix(statistics StatisticsPort) *ProcessMatrixUseCase {
	return &ProcessMatrixUseCase{statistics: statistics}
}

// Execute ejecuta el flujo completo del caso de uso.
func (uc *ProcessMatrixUseCase) Execute(ctx context.Context, authHeader string, rawMatrix [][]float64) (Output, error) {
	if err := matrix.Validate(rawMatrix); err != nil {
		return Output{}, err
	}

	rotated := matrix.Rotate90Clockwise(rawMatrix)
	factor, err := matrix.FactorizeQR(rawMatrix)
	if err != nil {
		return Output{}, fmt.Errorf("factorizar QR: %w", err)
	}

	result, err := uc.statistics.Statistics(ctx, authHeader, []stats.Matrix{
		{Name: "rotated", Values: rotated},
		{Name: "q", Values: factor.Q},
		{Name: "r", Values: factor.R},
	})
	if err != nil {
		return Output{}, &UpstreamError{Err: err}
	}

	return Output{
		Original:   rawMatrix,
		Rotated:    rotated,
		Rotation:   "90_clockwise",
		QR:         factor,
		Statistics: result,
	}, nil
}
