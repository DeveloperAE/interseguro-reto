package config

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret-de-al-menos-32-caracteres!!"

func TestFromEnv(t *testing.T) {
	t.Run("carga valores por defecto con secretos obligatorios", func(t *testing.T) {
		t.Setenv("JWT_SECRET", testSecret)
		t.Setenv("DEMO_PASSWORD", "pass123")
		t.Setenv("PORT", "")
		t.Setenv("DEMO_USERNAME", "")
		t.Setenv("NODE_API_URL", "")
		t.Setenv("CORS_ORIGIN", "")
		t.Setenv("TRUST_PROXY", "")

		cfg, err := FromEnv()
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if cfg.Port != "8080" || cfg.DemoUsername != "demo" || cfg.NodeAPIURL != "http://localhost:3000" || cfg.CORSOrigin != "*" || cfg.TrustProxy {
			t.Fatalf("defaults inesperados: %+v", cfg)
		}
		if cfg.TokenTTL != time.Hour || cfg.LoginMaxAttempts != 10 || cfg.LoginWindow != 15*time.Minute {
			t.Fatalf("límites inesperados: %+v", cfg)
		}
	})

	t.Run("respeta variables personalizadas", func(t *testing.T) {
		t.Setenv("JWT_SECRET", testSecret)
		t.Setenv("DEMO_PASSWORD", "pwd")
		t.Setenv("PORT", "9090")
		t.Setenv("DEMO_USERNAME", "admin")
		t.Setenv("NODE_API_URL", "http://api-node:3000")
		t.Setenv("CORS_ORIGIN", "https://example.com")

		cfg, err := FromEnv()
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if cfg.Port != "9090" || cfg.DemoUsername != "admin" || cfg.NodeAPIURL != "http://api-node:3000" || cfg.CORSOrigin != "https://example.com" {
			t.Fatalf("config inesperada: %+v", cfg)
		}
	})

	t.Run("TRUST_PROXY solo se activa con true o 1", func(t *testing.T) {
		t.Setenv("JWT_SECRET", testSecret)
		t.Setenv("DEMO_PASSWORD", "pwd")
		cases := map[string]bool{"true": true, "TRUE": true, "1": true, " true ": true, "": false, "false": false, "0": false, "yes": false}
		for value, want := range cases {
			t.Setenv("TRUST_PROXY", value)
			cfg, err := FromEnv()
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if cfg.TrustProxy != want {
				t.Fatalf("TRUST_PROXY=%q: %v, se esperaba %v", value, cfg.TrustProxy, want)
			}
		}
	})

	t.Run("falla si falta JWT_SECRET", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")
		t.Setenv("DEMO_PASSWORD", "pwd")
		if _, err := FromEnv(); err == nil {
			t.Fatal("se esperaba error por falta de JWT_SECRET")
		}
	})

	t.Run("rechaza JWT_SECRET de menos de 32 caracteres", func(t *testing.T) {
		t.Setenv("JWT_SECRET", strings.Repeat("a", 31))
		t.Setenv("DEMO_PASSWORD", "pwd")
		_, err := FromEnv()
		if err == nil || err.Error() != "JWT_SECRET debe tener al menos 32 caracteres" {
			t.Fatalf("error inesperado: %v", err)
		}

		t.Setenv("JWT_SECRET", strings.Repeat("a", 32))
		if _, err := FromEnv(); err != nil {
			t.Fatalf("32 caracteres debería aceptarse: %v", err)
		}
	})

	t.Run("falla si falta DEMO_PASSWORD", func(t *testing.T) {
		t.Setenv("JWT_SECRET", testSecret)
		t.Setenv("DEMO_PASSWORD", "")
		if _, err := FromEnv(); err == nil {
			t.Fatal("se esperaba error por falta de DEMO_PASSWORD")
		}
	})
}
