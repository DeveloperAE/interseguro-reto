package usecase

import (
	"context"
	"errors"
	"testing"

	"api-go/internal/matrix"
	"api-go/internal/stats"
)

type stubStatisticsPort struct {
	gotAuth     string
	gotMatrices []stats.Matrix
	result      stats.Result
	err         error
}

func (s *stubStatisticsPort) Statistics(_ context.Context, authorization string, matrices []stats.Matrix) (stats.Result, error) {
	s.gotAuth = authorization
	s.gotMatrices = matrices
	return s.result, s.err
}

func TestProcessMatrixUseCase_Execute(t *testing.T) {
	stub := &stubStatisticsPort{
		result: stats.Result{
			Max:               6,
			Min:               1,
			Average:           3.5,
			Sum:               21,
			HasDiagonalMatrix: false,
		},
	}
	uc := NewProcessMatrix(stub)

	out, err := uc.Execute(context.Background(), "Bearer token-123", [][]float64{{1, 2, 3}, {4, 5, 6}})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if out.Rotation != "90_clockwise" {
		t.Fatalf("rotación inesperada: %s", out.Rotation)
	}
	if stub.gotAuth != "Bearer token-123" {
		t.Fatalf("auth no propagado: %s", stub.gotAuth)
	}
	if len(stub.gotMatrices) != 3 || stub.gotMatrices[0].Name != "rotated" || stub.gotMatrices[1].Name != "q" || stub.gotMatrices[2].Name != "r" {
		t.Fatalf("matrices enviadas inesperadas: %+v", stub.gotMatrices)
	}
	if out.QR.ReconstructionError > 1e-8 {
		t.Fatalf("error de reconstrucción QR alto: %g", out.QR.ReconstructionError)
	}
	if out.Statistics.Sum != 21 {
		t.Fatalf("estadísticas inesperadas: %+v", out.Statistics)
	}
}

func TestProcessMatrixUseCase_ValidationError(t *testing.T) {
	uc := NewProcessMatrix(&stubStatisticsPort{})
	_, err := uc.Execute(context.Background(), "Bearer token", [][]float64{{1, 2}, {3}})
	var valErr *matrix.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("se esperaba ValidationError, se obtuvo: %v", err)
	}
}

func TestProcessMatrixUseCase_UpstreamError(t *testing.T) {
	cause := errors.New("connection refused")
	uc := NewProcessMatrix(&stubStatisticsPort{err: cause})
	_, err := uc.Execute(context.Background(), "Bearer token", [][]float64{{1, 2}, {3, 4}})
	var upErr *UpstreamError
	if !errors.As(err, &upErr) {
		t.Fatalf("se esperaba UpstreamError, se obtuvo: %v", err)
	}
	if !errors.Is(err, cause) || upErr.Error() == "" {
		t.Fatalf("unwrap o mensaje inválido: %v", upErr)
	}
}
