# Requirement Splitting

项目需求自动拆分与验收追踪系统。

这个项目目标是做一个自建 Web 管理后台，帮助负责人、项目经理/产品、研发、测试、验收人围绕同一套项目数据完成：

- 原始需求录入
- AI 分阶段拆分需求
- 开发任务生成
- 测试用例生成
- AI 辅助测试
- 人工复核测试结论
- 验收检查项管理
- 缺陷流转
- 需求变更影响分析
- 项目健康度感知

第一版后端使用 Go 实现，AI 能力先通过 deterministic stub provider 跑通业务闭环，后续再替换为真实模型供应商。

## Current Status

当前已实现：

- Go API 服务骨架
- PostgreSQL Docker Compose 配置
- 数据库初始化迁移
- 项目 CRUD 基础接口
- 原始需求录入接口
- AI 草稿生成接口
- AI stub provider
- AI 草稿结构化校验
- AI 草稿保存到 PostgreSQL
- 单元测试覆盖项目服务和 AI 草稿服务

当前还未实现：

- AI 草稿发布为正式计划
- 任务看板
- 测试用例确认与 AI 测试执行
- 缺陷流转接口
- 需求变更流程
- 健康度计算
- Web 前端
- 真实 AI provider

## Tech Stack

- Go 1.25+
- Chi HTTP router
- PostgreSQL 16
- pgx
- Docker Compose
- Go test

## Project Structure

```text
cmd/
  api/          API server entrypoint
  migrate/      simple migration runner
internal/
  ai/           AI task interface and stub provider
  config/       environment config
  domain/       models and enums
  http/         router and handlers
  repository/   PostgreSQL and in-memory repositories
  service/      business services
migrations/     database migrations
docs/           design and implementation planning docs
```

## Local Setup

### 1. Start PostgreSQL

```bash
docker compose up -d postgres
```

Check status:

```bash
docker compose ps
```

Expected status: `healthy`.

### 2. Run Database Migration

```bash
go run ./cmd/migrate up
```

Rollback:

```bash
go run ./cmd/migrate down
```

### 3. Start API

```bash
go run ./cmd/api
```

Default address:

```text
http://localhost:8080
```

Health check:

```bash
curl http://localhost:8080/healthz
```

Expected:

```json
{"status":"ok"}
```

## API Quick Test

Create a project:

```bash
curl -s -X POST http://localhost:8080/api/projects \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "AI项目经理系统",
    "objective": "自动拆分需求并跟踪验收",
    "scope": "第一版闭环"
  }'
```

List projects:

```bash
curl -s http://localhost:8080/api/projects
```

Add raw requirement:

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/requirements \
  -H 'Content-Type: application/json' \
  -d '{
    "content": "我要做一个项目需求自动拆分的系统，AI 写测试用例，也能辅助测试，人工复核。"
  }'
```

Generate AI split draft:

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/ai/split-requirement \
  -H 'Content-Type: application/json' \
  -d '{
    "requirement_id": "<requirement_id>",
    "content": "我要做一个项目需求自动拆分的系统，AI 写测试用例，也能辅助测试，人工复核。"
  }'
```

List AI drafts:

```bash
curl -s http://localhost:8080/api/projects/<project_id>/ai-drafts
```

## Tests

Run all tests:

```bash
go test ./...
```

## Environment Variables

See `.env.example`.

```bash
APP_ADDR=:8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/reqsplit?sslmode=disable
AI_PROVIDER=stub
AI_API_KEY=
```

## Design Docs

- `docs/superpowers/specs/2026-05-15-project-requirement-splitting-design.md`
- `docs/superpowers/plans/2026-05-15-project-requirement-splitting-go-implementation.md`

## GitHub Push Notes

GitHub no longer supports password authentication for Git operations.

If HTTPS push fails with invalid username or token, use GitHub CLI:

```bash
gh auth login
gh auth setup-git
git push -u origin main
```

Or use a Personal Access Token as the password when Git prompts for credentials.
