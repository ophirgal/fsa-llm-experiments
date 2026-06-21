package main

import (
	"log"

	"fsa-boilerplate/backend/api"
	"fsa-boilerplate/backend/config"
	"fsa-boilerplate/backend/dal"
)

func main() {
	cfg := config.Load()

	db, err := dal.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := dal.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	r := api.New(db)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
