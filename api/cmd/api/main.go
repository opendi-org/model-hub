package main

import (
	"log"
	"net/http"

	"opendi.org/model-hub/api/internal/config"
	"opendi.org/model-hub/api/internal/database"
	"opendi.org/model-hub/api/internal/routes"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DSN())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	// TODO: register routes and use db
	handler := routes.NewRouter(cfg, db)

	srv := &http.Server{
		Addr:    ":" + cfg.ModelHubPort,
		Handler: handler,
	}
	log.Printf("api listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}
