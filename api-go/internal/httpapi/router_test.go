package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-go/internal/config"
	"api-go/internal/stats"
)

func testConfig() config.Config {
	return config.Config{
		JWTSecret:        "test-secret-de-al-menos-32-caracteres!!",
		DemoUsername:     "demo",
		DemoPassword:     "demo123",
		CORSOrigin:       "*",
		TokenTTL:         time.Hour,
		LoginMaxAttempts: 10,
		LoginWindow:      time.Minute,
	}
}

func TestLoginRateLimit(t *testing.T) {
	cfg := testConfig()
	cfg.LoginMaxAttempts = 2
	app := New(cfg, stats.NewHTTPClient("http://127.0.0.1:1", time.Second))

	for i := 0; i < cfg.LoginMaxAttempts; i++ {
		status, _ := doJSON(t, app, http.MethodPost, "/v1/auth/login", `{"username":"demo","password":"no"}`, "")
		if status != http.StatusUnauthorized {
			t.Fatalf("intento %d: %d", i+1, status)
		}
	}

	status, payload := doJSON(t, app, http.MethodPost, "/v1/auth/login", `{"username":"demo","password":"demo123"}`, "")
	if status != http.StatusTooManyRequests {
		t.Fatalf("se esperaba 429, llegó %d: %s", status, payload)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "demasiados intentos de ingreso, intenta más tarde" {
		t.Fatalf("mensaje inesperado: %q", body.Error)
	}

	status, _ = doJSON(t, app, http.MethodGet, "/health", "", "")
	if status != http.StatusOK {
		t.Fatalf("el límite no debe afectar /health: %d", status)
	}
}

func TestLoginRateLimitUsesDefaultsWhenUnset(t *testing.T) {
	cfg := testConfig()
	cfg.LoginMaxAttempts = 0
	cfg.LoginWindow = 0
	app := New(cfg, stats.NewHTTPClient("http://127.0.0.1:1", time.Second))

	for i := 0; i < config.DefaultLoginMaxAttempts; i++ {
		status, _ := doJSON(t, app, http.MethodPost, "/v1/auth/login", `{"username":"demo","password":"no"}`, "")
		if status != http.StatusUnauthorized {
			t.Fatalf("intento %d: %d", i+1, status)
		}
	}
	status, _ := doJSON(t, app, http.MethodPost, "/v1/auth/login", `{"username":"demo","password":"no"}`, "")
	if status != http.StatusTooManyRequests {
		t.Fatalf("se esperaba 429 tras %d intentos, llegó %d", config.DefaultLoginMaxAttempts, status)
	}
}

func TestLoginRateLimitWithTrustProxyIsPerForwardedIP(t *testing.T) {
	cfg := testConfig()
	cfg.LoginMaxAttempts = 2
	cfg.TrustProxy = true
	app := New(cfg, stats.NewHTTPClient("http://127.0.0.1:1", time.Second))

	assertStatuses(t, loginStatuses(t, app, "203.0.113.10", 3), http.StatusUnauthorized, http.StatusUnauthorized, http.StatusTooManyRequests)
	assertStatuses(t, loginStatuses(t, app, "203.0.113.20", 2), http.StatusUnauthorized, http.StatusUnauthorized)
	// La IP que agrega el proxy va al final; cambiar la primera no evita el límite.
	assertStatuses(t, loginStatuses(t, app, "198.51.100.1, 203.0.113.10", 1), http.StatusTooManyRequests)
}

func TestLoginRateLimitWithoutTrustProxySharesLimit(t *testing.T) {
	cfg := testConfig()
	cfg.LoginMaxAttempts = 2
	app := New(cfg, stats.NewHTTPClient("http://127.0.0.1:1", time.Second))

	assertStatuses(t, loginStatuses(t, app, "203.0.113.10", 2), http.StatusUnauthorized, http.StatusUnauthorized)
	assertStatuses(t, loginStatuses(t, app, "203.0.113.20", 1), http.StatusTooManyRequests)
}

func loginStatuses(t *testing.T, app *fiber.App, forwardedFor string, attempts int) []int {
	t.Helper()
	statuses := make([]int, 0, attempts)
	for i := 0; i < attempts; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"username":"demo","password":"no"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", forwardedFor)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		statuses = append(statuses, resp.StatusCode)
	}
	return statuses
}

func assertStatuses(t *testing.T, got []int, want ...int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("statuses %v, se esperaba %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("statuses %v, se esperaba %v", got, want)
		}
	}
}

func TestProcessFlow(t *testing.T) {
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/statistics" {
			t.Errorf("ruta inesperada: %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Error("falta el token reenviado a node")
		}
		var body struct {
			Matrices []stats.Matrix `json:"matrices"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.Matrices) != 3 || body.Matrices[0].Name != "rotated" || body.Matrices[1].Name != "q" || body.Matrices[2].Name != "r" {
			t.Fatalf("payload inesperado: %+v", body.Matrices)
		}
		want := [][]float64{{4, 1}, {5, 2}, {6, 3}}
		if !same(body.Matrices[0].Values, want) {
			t.Fatalf("matriz rotada enviada: %v", body.Matrices[0].Values)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats.Result{
			Max: 6, Min: 1, Average: 3.5, Sum: 21,
			Matrices: []stats.Detail{
				{Name: "rotated", Rows: 3, Cols: 2},
				{Name: "q", Rows: 2, Cols: 2},
				{Name: "r", Rows: 2, Cols: 3},
			},
		})
	}))
	defer node.Close()

	app := New(testConfig(), stats.NewHTTPClient(node.URL, time.Second))
	token := obtainToken(t, app)

	status, payload := doJSON(t, app, http.MethodPost, "/v1/matrices/process", `{
		"matrix": [[1, 2, 3], [4, 5, 6]]
	}`, token)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, payload)
	}

	var body struct {
		Rotation string      `json:"rotation"`
		Rotated  [][]float64 `json:"rotated"`
		QR       struct {
			Q                   [][]float64 `json:"q"`
			R                   [][]float64 `json:"r"`
			ReconstructionError float64     `json:"reconstructionError"`
		} `json:"qr"`
		Statistics stats.Result `json:"statistics"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	if body.Rotation != "90_clockwise" || !same(body.Rotated, [][]float64{{4, 1}, {5, 2}, {6, 3}}) {
		t.Fatalf("rotación inesperada: %+v", body)
	}
	if len(body.QR.Q) == 0 || len(body.QR.R) == 0 || body.QR.ReconstructionError > 1e-8 {
		t.Fatalf("qr inesperado: %+v", body.QR)
	}
	if body.Statistics.Sum != 21 || body.Statistics.Max != 6 {
		t.Fatalf("estadísticas: %+v", body.Statistics)
	}
}

func TestProcessRejectsBadInput(t *testing.T) {
	app := New(testConfig(), stats.NewHTTPClient("http://127.0.0.1:1", time.Second))
	token := obtainToken(t, app)

	status, _ := doJSON(t, app, http.MethodPost, "/v1/matrices/process", `{"matrix":[[1,2],[3]]}`, token)
	if status != http.StatusBadRequest {
		t.Fatalf("status %d", status)
	}

	status, _ = doJSON(t, app, http.MethodPost, "/v1/matrices/process", `{"matrix":[[1,2],[3,4]]}`, "")
	if status != http.StatusUnauthorized {
		t.Fatalf("sin token: %d", status)
	}

	status, _ = doJSON(t, app, http.MethodPost, "/v1/auth/login", `{"username":"demo","password":"no"}`, "")
	if status != http.StatusUnauthorized {
		t.Fatalf("login: %d", status)
	}

	status, _ = doJSON(t, app, http.MethodPost, "/v1/auth/login", `{bad-json`, "")
	if status != http.StatusBadRequest {
		t.Fatalf("login bad json: %d", status)
	}

	status, _ = doJSON(t, app, http.MethodPost, "/v1/matrices/process", `{bad-json`, token)
	if status != http.StatusBadRequest {
		t.Fatalf("process bad json: %d", status)
	}

	status, _ = doJSON(t, app, http.MethodPost, "/v1/matrices/process", `{"matrix":[[1]]}`, "token-invalido")
	if status != http.StatusUnauthorized {
		t.Fatalf("token inválido: %d", status)
	}

	status, _ = doJSON(t, app, http.MethodGet, "/health", "", "")
	if status != http.StatusOK {
		t.Fatalf("health: %d", status)
	}
}

func TestProcessReturnsBadGatewayWhenNodeFails(t *testing.T) {
	app := New(testConfig(), stats.NewHTTPClient("http://127.0.0.1:1", 200*time.Millisecond))
	token := obtainToken(t, app)
	status, payload := doJSON(t, app, http.MethodPost, "/v1/matrices/process", `{"matrix":[[1,2],[3,4]]}`, token)
	if status != http.StatusBadGateway {
		t.Fatalf("status %d: %s", status, payload)
	}
}

func obtainToken(t *testing.T, app *fiber.App) string {
	t.Helper()
	status, payload := doJSON(t, app, http.MethodPost, "/v1/auth/login", `{"username":"demo","password":"demo123"}`, "")
	if status != http.StatusOK {
		t.Fatalf("login %d: %s", status, payload)
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	if body.Token == "" {
		t.Fatal("token vacío")
	}
	return body.Token
}

func doJSON(t *testing.T, app *fiber.App, method, path, raw, token string) (int, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, payload
}

func same(a, b [][]float64) bool {
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
