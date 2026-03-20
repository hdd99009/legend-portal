package main

import (
	"log"

	"legend-portal/internal/app"
	"legend-portal/internal/config"
	"legend-portal/internal/repository"
	"legend-portal/internal/service"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	repo, err := repository.NewSQLiteRepository(cfg.SQLite.Path)
	if err != nil {
		log.Fatalf("open sqlite failed: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate("migrations/001_init.sql"); err != nil {
		log.Fatalf("run migrations failed: %v", err)
	}

	renderer, err := app.NewTemplateRenderer(".")
	if err != nil {
		log.Fatalf("parse templates failed: %v", err)
	}

	siteService := service.NewSiteService(repo)
	adminService := service.NewAdminService(repo)
	server := app.NewServer(renderer, siteService, adminService, cfg.App.SessionSecret)

	log.Printf("server started at %s", cfg.App.Addr)
	if err := server.Start(cfg.App.Addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
