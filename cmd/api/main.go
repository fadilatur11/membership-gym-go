package main

import (
	"log"

	"membership-gym/config"
	"membership-gym/database"
	"membership-gym/internal/app"
)

// @title Gym Management SaaS API
// @version 1.0
// @description Gym Management SaaS backend API.
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token dengan format: Bearer {access_token}
func main() {
	cfg := config.Load()
	db, err := database.NewPostgresPool(cfg)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redisClient.Close()

	router := app.New(cfg, db, redisClient)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
