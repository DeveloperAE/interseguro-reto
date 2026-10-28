package main

import (
	"log"
	"time"

	"api-go/internal/config"
	"api-go/internal/httpapi"
	"api-go/internal/stats"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatal(err)
	}
	client := stats.NewHTTPClient(cfg.NodeAPIURL, 5*time.Second)
	app := httpapi.New(cfg, client)
	log.Printf("api-go escuchando en :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
