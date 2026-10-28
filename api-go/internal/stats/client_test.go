package stats

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClient_Statistics_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/statistics" || r.Method != http.MethodPost {
			t.Fatalf("solicitud inesperada: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-jwt" {
			t.Fatalf("header Authorization inesperado: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Result{
			Max:               9,
			Min:               -5,
			Average:           0.666,
			Sum:               6,
			HasDiagonalMatrix: true,
			Matrices: []Detail{
				{Name: "rotated", IsDiagonal: false, Rows: 3, Cols: 3},
			},
		})
	}))
	defer srv.Close()

	client := NewHTTPClient(srv.URL, time.Second)
	res, err := client.Statistics(context.Background(), "Bearer test-jwt", []Matrix{
		{Name: "rotated", Values: [][]float64{{1, 0}, {0, 2}}},
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !res.HasDiagonalMatrix || res.Max != 9 || res.Min != -5 {
		t.Fatalf("resultado inesperado: %+v", res)
	}
}

func TestHTTPClient_Statistics_UpstreamStatusErrors(t *testing.T) {
	t.Run("error json de node", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"matriz inválida"}`))
		}))
		defer srv.Close()

		client := NewHTTPClient(srv.URL, time.Second)
		_, err := client.Statistics(context.Background(), "Bearer tok", []Matrix{{Name: "m", Values: [][]float64{{1}}}})
		if err == nil {
			t.Fatal("se esperaba error")
		}
	})

	t.Run("error sin cuerpo json", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer srv.Close()

		client := NewHTTPClient(srv.URL, time.Second)
		_, err := client.Statistics(context.Background(), "Bearer tok", []Matrix{{Name: "m", Values: [][]float64{{1}}}})
		if err == nil {
			t.Fatal("se esperaba error")
		}
	})

	t.Run("respuesta 200 con json malformado", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{not-json`))
		}))
		defer srv.Close()

		client := NewHTTPClient(srv.URL, time.Second)
		_, err := client.Statistics(context.Background(), "Bearer tok", []Matrix{{Name: "m", Values: [][]float64{{1}}}})
		if err == nil {
			t.Fatal("se esperaba error de decodificación")
		}
	})
}

func TestHTTPClient_Statistics_NetworkError(t *testing.T) {
	client := NewHTTPClient("http://127.0.0.1:1", 100*time.Millisecond)
	_, err := client.Statistics(context.Background(), "Bearer tok", []Matrix{{Name: "m", Values: [][]float64{{1}}}})
	if err == nil {
		t.Fatal("se esperaba error de red")
	}
}
