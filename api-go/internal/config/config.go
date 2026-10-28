// Package config carga la configuración de la API en Go desde el entorno.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// MinJWTSecretLength es la longitud mínima aceptada para el secreto HS256.
const MinJWTSecretLength = 32

// Límites por defecto del login por IP.
const (
	DefaultLoginMaxAttempts = 10
	DefaultLoginWindow      = 15 * time.Minute
)

// Config es la configuración de proceso. No tiene valores secretos por defecto.
type Config struct {
	Port             string
	JWTSecret        string
	DemoUsername     string
	DemoPassword     string
	NodeAPIURL       string
	CORSOrigin       string
	TrustProxy       bool
	TokenTTL         time.Duration
	LoginMaxAttempts int
	LoginWindow      time.Duration
}

// FromEnv exige JWT_SECRET (mínimo 32 caracteres) y DEMO_PASSWORD. El resto tiene valores de desarrollo.
func FromEnv() (Config, error) {
	cfg := Config{
		Port:             env("PORT", "8080"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		DemoUsername:     env("DEMO_USERNAME", "demo"),
		DemoPassword:     os.Getenv("DEMO_PASSWORD"),
		NodeAPIURL:       env("NODE_API_URL", "http://localhost:3000"),
		CORSOrigin:       env("CORS_ORIGIN", "*"),
		TrustProxy:       parseBool(os.Getenv("TRUST_PROXY")),
		TokenTTL:         time.Hour,
		LoginMaxAttempts: DefaultLoginMaxAttempts,
		LoginWindow:      DefaultLoginWindow,
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET es obligatorio")
	}
	if len(cfg.JWTSecret) < MinJWTSecretLength {
		return Config{}, fmt.Errorf("JWT_SECRET debe tener al menos %d caracteres", MinJWTSecretLength)
	}
	if cfg.DemoPassword == "" {
		return Config{}, fmt.Errorf("DEMO_PASSWORD es obligatorio")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1":
		return true
	}
	return false
}
