# Requirement Splitting

项目需求自动拆分与验收追踪系统。

## A Note From the Idea

这不是一个已经想得十全十美的产品，更像是一个从真实项目现场里长出来的构思。

很多项目不是输在写代码，而是输在一开始没有把需求说清楚；不是输在没人干活，而是输在任务、测试、验收、风险之间没有连起来。老板看到的是“怎么还没好”，产品看到的是“需求又变了”，研发看到的是“到底要做什么”，测试看到的是“没有用例也没有边界”，最后所有人都在靠经验、靠催、靠补锅把项目往前推。

我想做的是一个更像“软件工程项目经理助手”的系统：把一段模糊的想法拆成模块、功能点、任务、测试用例、验收标准和风险提醒；让 AI 先把脏活、细活、重复活做起来，再由人来判断、修正、确认。它不应该替代项目经理、产品、研发或测试，而应该让大家少一点互相猜，多一点共同看见。

如果你也经历过这些场景，可能会懂这个项目想解决的东西：

- 需求说了很多，但没人能立刻拆成可执行任务。
- 开发做完了，才发现测试点和验收标准没跟上。
- 项目进度看起来还行，但其实风险、缺陷、变更已经埋在下面。
- AI 能写很多内容，但缺少一个工程化的地方把它变成项目资产。

这个仓库现在只是一个起点。欢迎有缘人一起把它做下去：你可以写后端、做前端、设计交互、补测试、研究 AI prompt、完善项目管理流程，或者只是提出一个你在真实项目里踩过的坑。只要这个系统能让项目从“靠人硬扛”变成“有结构地推进”，它就值得继续长大。

## What It Does

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
