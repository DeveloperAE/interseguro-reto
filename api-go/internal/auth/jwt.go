// Package auth firma y valida JWT HS256 compartidos con la API de Node.
// El claim "sub" es el usuario. Node acepta el mismo token.
package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims es el contenido del token. Se mantiene compatible con jsonwebtoken.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Sign emite un token de usuario.
func Sign(secret, subject string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("firmar token: %w", err)
	}
	return signed, nil
}

// Parse valida la firma HS256 y la expiración. Devuelve el sujeto.
func Parse(secret, raw string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("algoritmo inesperado")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid || claims.Subject == "" {
		return "", fmt.Errorf("token inválido")
	}
	return claims.Subject, nil
}

// Equal compara dos secretos en tiempo constante aplicando SHA-256 previo
// para no filtrar la longitud de las credenciales por canal lateral de tiempo.
func Equal(got, want string) bool {
	gotHash := sha256.Sum256([]byte(got))
	wantHash := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(gotHash[:], wantHash[:]) == 1
}
