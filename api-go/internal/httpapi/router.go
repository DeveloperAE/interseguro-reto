// Package httpapi actúa como Adaptador Primario HTTP (Driving Adapter) en Fiber v2:
// gestiona transporte JSON, autenticación JWT y delega la lógica al caso de uso.
package httpapi

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"api-go/internal/auth"
	"api-go/internal/config"
	"api-go/internal/matrix"
	"api-go/internal/stats"
	"api-go/internal/usecase"
)

// New construye el adaptador HTTP en Fiber inyectando configuración y el puerto de estadísticas.
func New(cfg config.Config, statistics stats.Client) *fiber.App {
	processUC := usecase.NewProcessMatrix(statistics)

	app := fiber.New(fiber.Config{
		AppName:               "matrix-go-api",
		BodyLimit:             1 << 20,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           30 * time.Second,
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				code = fiberErr.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(logger.New(logger.Config{Format: "${time} ${status} ${method} ${path} ${latency}\n"}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigin,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,OPTIONS",
	}))
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		return c.Next()
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "api-go"})
	})

	v1 := app.Group("/v1")
	v1.Post("/auth/login", loginLimiter(cfg), login(cfg))

	protected := v1.Group("", requireJWT(cfg.JWTSecret))
	protected.Post("/matrices/process", process(processUC))

	return app
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type processRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

type processResponse struct {
	Original   [][]float64  `json:"original"`
	Rotated    [][]float64  `json:"rotated"`
	Rotation   string       `json:"rotation"`
	QR         qrResponse   `json:"qr"`
	Statistics stats.Result `json:"statistics"`
}

type qrResponse struct {
	Q                   [][]float64 `json:"q"`
	R                   [][]float64 `json:"r"`
	ReconstructionError float64     `json:"reconstructionError"`
}

func login(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body loginRequest
		if err := c.BodyParser(&body); err != nil {
			return writeError(c, fiber.StatusBadRequest, "el cuerpo debe ser JSON con username y password")
		}
		if !auth.Equal(body.Username, cfg.DemoUsername) || !auth.Equal(body.Password, cfg.DemoPassword) {
			return writeError(c, fiber.StatusUnauthorized, "credenciales inválidas")
		}
		token, err := auth.Sign(cfg.JWTSecret, body.Username, cfg.TokenTTL)
		if err != nil {
			log.Printf("login: %v", err)
			return writeError(c, fiber.StatusInternalServerError, "no se pudo emitir el token")
		}
		return c.JSON(fiber.Map{
			"token":     token,
			"expiresIn": int(cfg.TokenTTL / time.Second),
		})
	}
}

// loginLimiter cuenta intentos por IP en memoria del proceso: con varias réplicas cada una limita por separado.
func loginLimiter(cfg config.Config) fiber.Handler {
	maxAttempts := cfg.LoginMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = config.DefaultLoginMaxAttempts
	}
	window := cfg.LoginWindow
	if window <= 0 {
		window = config.DefaultLoginWindow
	}
	return limiter.New(limiter.Config{
		Max:          maxAttempts,
		Expiration:   window,
		KeyGenerator: clientIP(cfg.TrustProxy),
		LimitReached: func(c *fiber.Ctx) error {
			return writeError(c, fiber.StatusTooManyRequests, "demasiados intentos de ingreso, intenta más tarde")
		},
	})
}

// clientIP con trustProxy toma la última IP de X-Forwarded-For, la que agrega el proxy (un salto,
// igual que trust proxy 1 en Express). La primera la puede enviar el cliente y no sirve para limitar.
func clientIP(trustProxy bool) func(*fiber.Ctx) string {
	return func(c *fiber.Ctx) string {
		if trustProxy {
			if ips := c.IPs(); len(ips) > 0 {
				return ips[len(ips)-1]
			}
		}
		return c.IP()
	}
}

func process(uc *usecase.ProcessMatrixUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body processRequest
		if err := c.BodyParser(&body); err != nil {
			return writeError(c, fiber.StatusBadRequest, "el cuerpo debe ser JSON con el campo matrix")
		}

		out, err := uc.Execute(c.UserContext(), c.Get("Authorization"), body.Matrix)
		if err != nil {
			var valErr *matrix.ValidationError
			if errors.As(err, &valErr) {
				return writeError(c, fiber.StatusBadRequest, valErr.Message)
			}
			var upErr *usecase.UpstreamError
			if errors.As(err, &upErr) {
				log.Printf("estadisticas: %v", upErr)
				return writeError(c, fiber.StatusBadGateway, "no se pudieron calcular las estadísticas")
			}
			log.Printf("procesamiento: %v", err)
			return writeError(c, fiber.StatusInternalServerError, "no se pudo factorizar la matriz")
		}

		return c.JSON(processResponse{
			Original: out.Original,
			Rotated:  out.Rotated,
			Rotation: out.Rotation,
			QR: qrResponse{
				Q:                   out.QR.Q,
				R:                   out.QR.R,
				ReconstructionError: out.QR.ReconstructionError,
			},
			Statistics: out.Statistics,
		})
	}
}

func requireJWT(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return writeError(c, fiber.StatusUnauthorized, "se requiere un token Bearer")
		}
		if _, err := auth.Parse(secret, parts[1]); err != nil {
			return writeError(c, fiber.StatusUnauthorized, "token inválido o vencido")
		}
		return c.Next()
	}
}

func writeError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}
