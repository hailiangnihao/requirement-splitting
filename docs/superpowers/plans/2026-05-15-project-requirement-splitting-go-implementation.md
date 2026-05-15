# Project Requirement Splitting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first usable version of a self-hosted Web system that turns raw project requirements into AI-assisted plans, tasks, test cases, acceptance items, defects, change impact analysis, and progress health tracking.

**Architecture:** Use a Go backend as the source of truth and API layer, PostgreSQL for relational project data, and a Web frontend for role-specific project management views. AI is isolated behind an internal service interface so the product calls business tasks such as requirement splitting, test generation, AI-assisted test execution, and risk explanation without coupling to one model provider.

**Tech Stack:** Go 1.22+, Gin or Chi HTTP router, PostgreSQL, sqlc or GORM, golang-migrate, Vite + Vue 3 + TypeScript, OpenAPI, Docker Compose, Go test, Vitest/Playwright for frontend verification.

---

## Implementation Strategy

This spec is larger than one coding task. The first delivery should be a working vertical slice, not the full mature PM system.

Build in this order:

1. Backend foundation and schema.
2. Project and requirement CRUD.
3. AI draft generation and publish flow using a stub AI provider first.
4. Task, test case, acceptance item, defect, and change workflows.
5. Health score and risk explanation.
6. Web UI for the core single-project loop.
7. Replace stub AI with a configurable model provider.

The first implementation should keep AI provider calls replaceable. Tests should cover service logic before UI polish.

## Repository Layout

Create:

- `go.mod`: Go module definition.
- `cmd/api/main.go`: API server entrypoint.
- `internal/config/config.go`: Environment config loading.
- `internal/http/router.go`: Route registration and middleware.
- `internal/http/handlers/*.go`: HTTP handlers grouped by domain.
- `internal/domain/*.go`: Domain models, enums, and validation helpers.
- `internal/service/*.go`: Business workflows.
- `internal/repository/*.go`: PostgreSQL persistence interfaces and implementations.
- `internal/ai/*.go`: AI task interfaces, prompts, response schemas, provider adapters, and stub provider.
- `internal/health/*.go`: Project health scoring rules.
- `migrations/*.sql`: Database migrations.
- `web/`: Vue admin frontend.
- `docker-compose.yml`: Local PostgreSQL and app dependencies.
- `.env.example`: Local config template.
- `docs/api/openapi.yaml`: Public API contract.

Keep business rules in `internal/service` and `internal/health`; handlers should only parse input, call services, and shape responses.

## Data Model

Core tables:

- `users`: simple local users and role metadata.
- `projects`: project root, status, owner, health level.
- `project_members`: project role assignments.
- `requirements`: raw requirement text and uploaded document metadata.
- `ai_drafts`: AI generated drafts, task type, provider metadata, JSON result, validation errors, status.
- `plan_versions`: published formal plan versions.
- `modules`: project modules.
- `milestones`: project phases.
- `feature_points`: traceability center for tasks, tests, and acceptance.
- `dev_tasks`: development tasks and dependencies.
- `test_cases`: AI/manual test cases, execution state, AI result, human review state.
- `test_runs`: individual AI or human execution records and evidence.
- `acceptance_items`: business acceptance checks.
- `defects`: defect lifecycle and acceptance blocking flag.
- `change_requests`: requirement change workflow and AI impact analysis.
- `risk_items`: rule and AI generated risks or gaps.
- `ai_invocations`: AI call log for traceability.

Use explicit enums for statuses. Store AI structured results as JSONB, but publish normalized data into formal tables.

## API Surface

Minimum endpoints:

- `POST /api/projects`
- `GET /api/projects`
- `GET /api/projects/:id`
- `POST /api/projects/:id/requirements`
- `POST /api/projects/:id/ai/split-requirement`
- `GET /api/projects/:id/ai-drafts`
- `POST /api/ai-drafts/:id/publish`
- `GET /api/projects/:id/plan`
- `PATCH /api/dev-tasks/:id/status`
- `POST /api/projects/:id/test-cases/generate`
- `POST /api/test-cases/:id/confirm`
- `POST /api/test-cases/:id/ai-run`
- `POST /api/test-runs/:id/review`
- `POST /api/defects`
- `PATCH /api/defects/:id/status`
- `POST /api/projects/:id/change-requests`
- `POST /api/change-requests/:id/analyze`
- `POST /api/change-requests/:id/apply`
- `GET /api/projects/:id/health`

## Task 1: Backend Scaffold

**Files:**
- Create: `go.mod`
- Create: `cmd/api/main.go`
- Create: `internal/config/config.go`
- Create: `internal/http/router.go`
- Create: `.env.example`
- Create: `docker-compose.yml`

- [x] **Step 1: Initialize Go module**

Run: `go mod init requirement-splitting`

Expected: `go.mod` exists.

- [x] **Step 2: Add HTTP dependencies**

Run: `go get github.com/go-chi/chi/v5 github.com/jackc/pgx/v5/pgxpool`

Expected: dependencies are added to `go.mod`.

- [x] **Step 3: Add config loader**

Implement `internal/config/config.go` with:

```go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Addr        string
	DatabaseURL string
	AIProvider  string
	AIAPIKey    string
}

func Load() Config {
	return Config{
		Addr:        env("APP_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/reqsplit?sslmode=disable"),
		AIProvider:  env("AI_PROVIDER", "stub"),
		AIAPIKey:    os.Getenv("AI_API_KEY"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func EnvInt(key string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return v
}
```

- [x] **Step 4: Add health route**

Implement `internal/http/router.go` with `GET /healthz` returning `{"status":"ok"}`.

- [x] **Step 5: Start API**

Implement `cmd/api/main.go` to load config, create router, and run `http.ListenAndServe`.

- [x] **Step 6: Verify**

Run: `go test ./...`

Expected: PASS.

Run: `go run ./cmd/api`

Expected: server listens on `:8080` and `/healthz` returns ok.

## Task 2: Database Migrations and Domain Enums

**Files:**
- Create: `migrations/000001_init.up.sql`
- Create: `migrations/000001_init.down.sql`
- Create: `internal/domain/enums.go`
- Create: `internal/domain/models.go`

- [x] **Step 1: Write migration**

Create tables listed in the Data Model section. Use UUID primary keys, `created_at`, `updated_at`, and foreign keys for project-scoped data.

- [x] **Step 2: Add status enums in Go**

Define constants for:

- Project status.
- AI draft status.
- Task status.
- Test case confirmation status.
- Test run review status.
- Defect status.
- Change request status.
- Health level.

- [x] **Step 3: Add migration command**

Use `golang-migrate/migrate` or a small migration runner. Prefer `golang-migrate` if CLI availability is acceptable.

- [x] **Step 4: Verify**

Run database locally:

`docker compose up -d postgres`

Run migrations.

Expected: all tables are created successfully.

Progress: `docker compose config`, `docker compose up -d postgres`, `go run ./cmd/migrate up`, and `go test ./...` pass.

## Task 3: Project and Requirement APIs

**Files:**
- Create: `internal/repository/project_repository.go`
- Create: `internal/service/project_service.go`
- Create: `internal/http/handlers/project_handler.go`
- Test: `internal/service/project_service_test.go`

- [x] **Step 1: Write service tests**

Cover:

- Creating project validates name.
- Creating project creates owner membership.
- Adding raw requirement requires non-empty content.

- [x] **Step 2: Implement repositories**

Add persistence for projects, members, and requirements.

- [x] **Step 3: Implement service**

Service methods:

- `CreateProject`
- `ListProjects`
- `GetProject`
- `AddRequirement`

- [x] **Step 4: Implement handlers**

Add endpoints:

- `POST /api/projects`
- `GET /api/projects`
- `GET /api/projects/:id`
- `POST /api/projects/:id/requirements`

- [x] **Step 5: Verify**

Run: `go test ./internal/service ./internal/http/...`

Expected: PASS.

## Task 4: AI Draft Service With Stub Provider

**Files:**
- Create: `internal/ai/types.go`
- Create: `internal/ai/stub_provider.go`
- Create: `internal/service/ai_draft_service.go`
- Create: `internal/repository/ai_draft_repository.go`
- Create: `internal/http/handlers/ai_handler.go`
- Test: `internal/service/ai_draft_service_test.go`

- [x] **Step 1: Define AI task interface**

```go
type TaskType string

const (
	TaskSplitRequirement TaskType = "split_requirement"
	TaskGenerateTestCases TaskType = "generate_test_cases"
	TaskExecuteAITest TaskType = "execute_ai_test"
	TaskAnalyzeChangeImpact TaskType = "analyze_change_impact"
	TaskExplainRisk TaskType = "explain_risk"
)

type Provider interface {
	Run(ctx context.Context, input TaskInput) (TaskOutput, error)
}
```

- [x] **Step 2: Write stub provider**

Return deterministic JSON containing modules, milestones, feature points, dev tasks, test cases, acceptance items, and risk items.

- [x] **Step 3: Validate AI output**

Require:

- Parseable JSON.
- Feature points exist before tasks/tests/acceptance link to them.
- Test cases include preconditions, steps, test data, and expected result.
- Acceptance items have pass criteria.

- [x] **Step 4: Save draft**

Save raw AI output, validation result, and status to `ai_drafts`.

- [x] **Step 5: Add endpoints**

- `POST /api/projects/:id/ai/split-requirement`
- `GET /api/projects/:id/ai-drafts`

- [x] **Step 6: Verify**

Run service tests.

Expected: invalid AI output stays unpublishable; valid stub output creates a draft.

## Task 5: Publish Draft to Formal Plan

**Files:**
- Create: `internal/service/plan_publish_service.go`
- Create: `internal/repository/plan_repository.go`
- Create: `internal/http/handlers/plan_handler.go`
- Test: `internal/service/plan_publish_service_test.go`

- [ ] **Step 1: Write publish tests**

Cover:

- Valid draft publishes modules, milestones, feature points, tasks, test cases, and acceptance items.
- Unvalidated draft cannot publish.
- Publishing creates a new `plan_versions` row.
- Test cases start as AI-generated but pending human confirmation.

- [ ] **Step 2: Implement transaction**

Publishing must be one database transaction.

- [ ] **Step 3: Add endpoint**

`POST /api/ai-drafts/:id/publish`

- [ ] **Step 4: Add plan query**

`GET /api/projects/:id/plan` returns module tree, tasks, tests, acceptance, and risks.

- [ ] **Step 5: Verify**

Run: `go test ./internal/service -run Publish`

Expected: PASS.

## Task 6: Test Case Confirmation and AI-Assisted Execution

**Files:**
- Create: `internal/service/test_service.go`
- Create: `internal/repository/test_repository.go`
- Create: `internal/http/handlers/test_handler.go`
- Test: `internal/service/test_service_test.go`

- [ ] **Step 1: Write tests**

Cover:

- AI-generated test case does not count toward formal coverage until confirmed.
- Confirming test case moves it into formal test plan.
- AI test execution creates a `test_runs` record with evidence.
- AI test run does not affect final result until human review.
- Human review can mark pass, fail, retest, or ignored.

- [ ] **Step 2: Implement confirmation**

Endpoint: `POST /api/test-cases/:id/confirm`

- [ ] **Step 3: Implement AI test execution**

Endpoint: `POST /api/test-cases/:id/ai-run`

Use AI provider task `execute_ai_test`. Stub provider returns deterministic evidence.

- [ ] **Step 4: Implement test run review**

Endpoint: `POST /api/test-runs/:id/review`

- [ ] **Step 5: Verify**

Run: `go test ./internal/service -run Test`

Expected: PASS.

## Task 7: Defect Workflow

**Files:**
- Create: `internal/service/defect_service.go`
- Create: `internal/repository/defect_repository.go`
- Create: `internal/http/handlers/defect_handler.go`
- Test: `internal/service/defect_service_test.go`

- [ ] **Step 1: Write status transition tests**

Allowed path:

`pending_confirm -> pending_fix -> fixing -> pending_regression -> regression_passed -> closed`

Also allow `rejected` from pending confirmation.

- [ ] **Step 2: Implement defect creation**

Support manual creation and creation from AI-generated defect draft.

- [ ] **Step 3: Implement status updates**

Endpoint: `PATCH /api/defects/:id/status`

- [ ] **Step 4: Verify acceptance blocking**

Blocking defects must appear in project health and acceptance views.

## Task 8: Requirement Change Workflow

**Files:**
- Create: `internal/service/change_service.go`
- Create: `internal/repository/change_repository.go`
- Create: `internal/http/handlers/change_handler.go`
- Test: `internal/service/change_service_test.go`

- [ ] **Step 1: Write tests**

Cover:

- Creating change request stores reason and content.
- AI impact analysis links affected modules, feature points, tasks, tests, acceptance, defects, and milestones.
- Applying accepted change creates new draft or formal version.
- Test case suggestions include add, modify, and deprecate actions.

- [ ] **Step 2: Implement endpoints**

- `POST /api/projects/:id/change-requests`
- `POST /api/change-requests/:id/analyze`
- `POST /api/change-requests/:id/apply`

- [ ] **Step 3: Verify**

Run: `go test ./internal/service -run Change`

Expected: PASS.

## Task 9: Health Score and Risk Explanation

**Files:**
- Create: `internal/health/scorer.go`
- Create: `internal/service/health_service.go`
- Create: `internal/http/handlers/health_handler.go`
- Test: `internal/health/scorer_test.go`

- [ ] **Step 1: Write scoring tests**

Cover:

- Unconfirmed test cases do not count toward coverage.
- Blocking defects lower health level.
- Missing acceptance items lower health level.
- Overdue tasks lower health level.
- Unhandled change impacts lower health level.

- [ ] **Step 2: Implement deterministic score**

Return:

- Numeric score.
- Health level: healthy, attention, risk, severe_risk.
- Triggered reasons.

- [ ] **Step 3: Add AI explanation**

Call AI provider task `explain_risk` using deterministic score context.

- [ ] **Step 4: Add endpoint**

`GET /api/projects/:id/health`

## Task 10: Web Frontend Scaffold

**Files:**
- Create: `web/package.json`
- Create: `web/src/main.ts`
- Create: `web/src/router.ts`
- Create: `web/src/api/client.ts`
- Create: `web/src/pages/*.vue`
- Create: `web/src/components/*.vue`

- [ ] **Step 1: Initialize Vite Vue**

Run inside `web`: `npm create vite@latest . -- --template vue-ts`

- [ ] **Step 2: Add routes**

Routes:

- `/projects`
- `/projects/:id/overview`
- `/projects/:id/requirements`
- `/projects/:id/tasks`
- `/projects/:id/tests`
- `/projects/:id/defects`
- `/projects/:id/changes`
- `/projects/:id/risks`

- [ ] **Step 3: Build API client**

Use typed request wrappers for backend endpoints.

- [ ] **Step 4: Implement usable screens**

Keep first UI dense and operational:

- Project list.
- Project overview.
- Requirement split and draft publish page.
- Task board.
- Test and acceptance page.
- Defect list.
- Change request page.
- Risk and health page.

- [ ] **Step 5: Verify**

Run: `npm run build`

Expected: PASS.

## Task 11: End-to-End Local Demo

**Files:**
- Modify: `README.md`
- Create: `docs/demo-script.md`

- [ ] **Step 1: Write README**

Include:

- Prerequisites.
- Environment variables.
- Database startup.
- Migration command.
- API startup.
- Web startup.

- [ ] **Step 2: Write demo script**

Demo path:

1. Create project.
2. Add raw requirement.
3. Generate AI split draft.
4. Publish plan.
5. Confirm test cases.
6. Run AI-assisted test.
7. Review result and create defect.
8. Move defect through regression.
9. Create change request.
10. Analyze impact.
11. View health score.

- [ ] **Step 3: Run full verification**

Run:

```bash
go test ./...
cd web && npm run build
```

Expected: all tests and build pass.

## Implementation Notes

- Start with stub AI provider. Do not block core workflow on live model integration.
- Keep AI output versioned and auditable.
- Never let AI directly overwrite formal plan data.
- Never let unreviewed AI test results count as final test or acceptance evidence.
- Health score must be deterministic; AI only explains the deterministic result.
- Use `GOCACHE` under a writable temp directory if local sandbox blocks Go cache writes.

## First Milestone Definition

The first milestone is complete when a user can run the app locally and complete this flow:

Raw requirement -> AI split draft -> publish plan -> confirm AI test cases -> AI-assisted test run -> human review -> defect workflow -> health score.

Requirement change, polished UI, and live AI provider can follow after this vertical slice if needed.
