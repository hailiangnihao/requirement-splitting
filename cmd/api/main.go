package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/ai"
	"requirement-splitting/internal/config"
	apphttp "requirement-splitting/internal/http"
	"requirement-splitting/internal/repository"
	"requirement-splitting/internal/service"
)

func main() {
	cfg := config.Load()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	ctx := context.Background()

	// 配置数据库连接池
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("parse database config: %v", err)
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
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
	aiDraftService := service.NewAIDraftService(aiDraftRepo, aiProvider)

	planRepo := repository.NewPGPlanRepository(pool)
	planPublishService := service.NewPlanPublishService(aiDraftRepo, planRepo)

	testRepo := repository.NewPGTestRepository(pool)
	testService := service.NewTestService(testRepo, aiProvider)

	defectRepo := repository.NewPGDefectRepository(pool)
	defectService := service.NewDefectService(defectRepo, testRepo, testService)

	changeRepo := repository.NewPGChangeRepository(pool)
	changeService := service.NewChangeService(changeRepo, planRepo, aiProvider)

	healthRepo := repository.NewPGHealthRepository(pool)
	healthService := service.NewHealthService(healthRepo, aiProvider)

	router := apphttp.NewRouter(projectService, aiDraftService, planPublishService, testService, defectService, changeService, healthService)

	// 配置 HTTP 服务器
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("api listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
