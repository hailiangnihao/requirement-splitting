package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"requirement-splitting/internal/config"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: go run ./cmd/migrate <up|down>")
	}

	direction := os.Args[1]
	if direction != "up" && direction != "down" {
		log.Fatalf("unsupported migration direction %q", direction)
	}

	cfg := config.Load()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	files, err := migrationFiles("migrations", direction)
	if err != nil {
		log.Fatalf("load migrations: %v", err)
	}

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("read migration %s: %v", file, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			log.Fatalf("execute migration %s: %v", file, err)
		}
		fmt.Printf("applied %s\n", file)
	}
}

func migrationFiles(dir, direction string) ([]string, error) {
	pattern := filepath.Join(dir, fmt.Sprintf("*.%s.sql", direction))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	if direction == "down" {
		for left, right := 0, len(files)-1; left < right; left, right = left+1, right-1 {
			files[left], files[right] = files[right], files[left]
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no %s migration files found in %s", direction, strings.TrimSuffix(pattern, filepath.Base(pattern)))
	}
	return files, nil
}
