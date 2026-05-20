package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/ai"
	"requirement-splitting/internal/config"
	apphttp "requirement-splitting/internal/http"
	"requirement-splitting/internal/repository"
	"requirement-splitting/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	projectRepo := repository.NewPGProjectRepository(pool)
	projectService := service.NewProjectService(projectRepo)
	aiDraftRepo := repository.NewPGAIDraftRepository(pool)
	aiDraftService := service.NewAIDraftService(aiDraftRepo, ai.NewStubProvider())

	planRepo := repository.NewPGPlanRepository(pool)
	planPublishService := service.NewPlanPublishService(aiDraftRepo, planRepo)

	router := apphttp.NewRouter(projectService, aiDraftService, planPublishService)

	log.Printf("api listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatalf("api server stopped: %v", err)
	}
}
