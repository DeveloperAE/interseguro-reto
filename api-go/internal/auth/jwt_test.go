package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignAndParse(t *testing.T) {
	secret := "test-secret-key"
	token, err := Sign(secret, "demo", time.Minute)
	if err != nil {
		t.Fatalf("Sign falló: %v", err)
	}
	sub, err := Parse(secret, token)
	if err != nil {
		t.Fatalf("Parse falló: %v", err)
	}
	if sub != "demo" {
		t.Fatalf("subject = %q, want %q", sub, "demo")
	}
}

func TestParseRejectsInvalidTokens(t *testing.T) {
	secret := "test-secret-key"

	t.Run("token vencido", func(t *testing.T) {
		expired, err := Sign(secret, "demo", -time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Parse(secret, expired); err == nil {
			t.Fatal("se esperaba error por token vencido")
		}
	})

	t.Run("secreto incorrecto", func(t *testing.T) {
		tok, err := Sign(secret, "demo", time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Parse("otro-secreto", tok); err == nil {
			t.Fatal("se esperaba error por secreto distinto")
		}
	})

	t.Run("algoritmo distinto a HS256", func(t *testing.T) {
		tok := jwt.NewWithClaims(jwt.SigningMethodHS384, Claims{
			Role: "user",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "demo",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			},
		})
		raw, err := tok.SignedString([]byte(secret))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Parse(secret, raw); err == nil {
			t.Fatal("se esperaba error por algoritmo inesperado")
		}
	})

	t.Run("subject vacío", func(t *testing.T) {
		tok, err := Sign(secret, "", time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Parse(secret, tok); err == nil {
			t.Fatal("se esperaba error por subject vacío")
		}
	})
}

func TestEqual(t *testing.T) {
	if !Equal("demo123", "demo123") {
		t.Fatal("debería coincidir cuando son iguales")
	}
	if Equal("demo123", "demo124") {
		t.Fatal("no debería coincidir con contenido distinto")
	}
	if Equal("corto", "mucho-mas-largo") {
		t.Fatal("no debería coincidir con longitudes distintas")
	}
}
