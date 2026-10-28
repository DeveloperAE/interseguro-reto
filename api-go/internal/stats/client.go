// Package stats es el cliente HTTP hacia la API de Node.
// Go no calcula máximo, mínimo, promedio, suma ni diagonal.
package stats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Matrix es una matriz con nombre, como la espera Node.
type Matrix struct {
	Name   string      `json:"name"`
	Values [][]float64 `json:"values"`
}

// Detail describe una matriz analizada.
type Detail struct {
	Name       string `json:"name"`
	IsDiagonal bool   `json:"isDiagonal"`
	Rows       int    `json:"rows"`
	Cols       int    `json:"cols"`
}

// Result es la respuesta de Node, reenviada sin recalcular.
type Result struct {
	Max               float64  `json:"max"`
	Min               float64  `json:"min"`
	Average           float64  `json:"average"`
	Sum               float64  `json:"sum"`
	HasDiagonalMatrix bool     `json:"hasDiagonalMatrix"`
	Matrices          []Detail `json:"matrices"`
}

// Client consulta las estadísticas.
type Client interface {
	Statistics(ctx context.Context, authorization string, matrices []Matrix) (Result, error)
}

// HTTPClient llama a POST /v1/statistics.
type HTTPClient struct {
	baseURL string
	http    *http.Client
}

// NewHTTPClient crea el cliente con un tiempo de espera acotado.
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: timeout},
	}
}

type requestBody struct {
	Matrices []Matrix `json:"matrices"`
}

type errorBody struct {
	Error string `json:"error"`
}

// Statistics reenvía el token del usuario. Node valida el mismo JWT.
func (c *HTTPClient) Statistics(ctx context.Context, authorization string, matrices []Matrix) (Result, error) {
	payload, err := json.Marshal(requestBody{Matrices: matrices})
	if err != nil {
		return Result{}, fmt.Errorf("armar solicitud: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/statistics", bytes.NewReader(payload))
	if err != nil {
		return Result{}, fmt.Errorf("crear solicitud: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorization)

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("llamar a node: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Result{}, fmt.Errorf("leer respuesta de node: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr errorBody
		_ = json.Unmarshal(body, &apiErr)
		if apiErr.Error == "" {
			apiErr.Error = resp.Status
		}
		return Result{}, fmt.Errorf("node respondió %d: %s", resp.StatusCode, apiErr.Error)
	}

	var result Result
	if err := json.Unmarshal(body, &result); err != nil {
		return Result{}, fmt.Errorf("decodificar estadísticas: %w", err)
	}
	return result, nil
}
