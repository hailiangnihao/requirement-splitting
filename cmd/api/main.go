package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

	// 初始化 AI Provider (根据环境变量动态切换)
	var aiProvider ai.Provider
	if os.Getenv("AI_PROVIDER") == "openai" {
		apiKey := os.Getenv("AI_API_KEY")
		apiURL := os.Getenv("AI_API_URL")
		model := os.Getenv("AI_MODEL")
		aiProvider = ai.NewOpenAIProvider(apiKey, apiURL, model)
		log.Printf("🤖 Using Real OpenAI (or compatible) Provider [Model: %s]\n", model)
	} else {
		aiProvider = ai.NewStubProvider()
		log.Println("🧸 Using Stub Provider (Mock)")
	}

	projectRepo := repository.NewPGProjectRepository(pool)
	projectService := service.NewProjectService(projectRepo)
	aiDraftRepo := repository.NewPGAIDraftRepository(pool)
	aiDraftService := service.NewAIDraftService(aiDraftRepo, aiProvider) // 替换为动态 Provider

	planRepo := repository.NewPGPlanRepository(pool)
	planPublishService := service.NewPlanPublishService(aiDraftRepo, planRepo)

	testRepo := repository.NewPGTestRepository(pool)
	testService := service.NewTestService(testRepo, aiProvider)

	defectRepo := repository.NewPGDefectRepository(pool)
	// 注意看这里，testService 直接作为 AITestRunner 接口的实现传给了 defectService！
	defectService := service.NewDefectService(defectRepo, testRepo, testService)

	changeRepo := repository.NewPGChangeRepository(pool)
	changeService := service.NewChangeService(changeRepo, planRepo, aiProvider)

	healthRepo := repository.NewPGHealthRepository(pool)
	healthService := service.NewHealthService(healthRepo, aiProvider)

	router := apphttp.NewRouter(projectService, aiDraftService, planPublishService, testService, defectService, changeService, healthService)

	log.Printf("api listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatalf("api server stopped: %v", err)
	}
}
