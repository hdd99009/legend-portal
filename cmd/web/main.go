package main

import (
	"log"

	"legend-portal/internal/app"
	"legend-portal/internal/config"
	"legend-portal/internal/repository"
	"legend-portal/internal/service"
	appstorage "legend-portal/internal/storage"
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

	var storage appstorage.FileStorage
	switch cfg.Storage.Driver {
	case "", "local":
		storage = appstorage.NewLocalStorage(cfg.Storage.LocalPath, cfg.Storage.PublicPrefix)
	default:
		log.Fatalf("unsupported storage driver: %s", cfg.Storage.Driver)
	}

	siteService := service.NewSiteService(repo)
	adminService := service.NewAdminService(repo, storage)
	server := app.NewServer(renderer, siteService, adminService, cfg.App.SessionSecret, storage)

	log.Printf("server started at %s", cfg.App.Addr)
	if err := server.Start(cfg.App.Addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
