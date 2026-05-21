# Requirement Splitting

项目需求自动拆分与验收追踪系统。

## A Note From the Idea

这不是一个已经想得十全十美的产品，更像是一个从真实项目现场里长出来的构思。

很多项目不是输在写代码，而是输在一开始没有把需求说清楚；不是输在没人干活，而是输在任务、测试、验收、风险之间没有连起来。老板看到的是”怎么还没好”，产品看到的是”需求又变了”，研发看到的是”到底要做什么”，测试看到的是”没有用例也没有边界”，最后所有人都在靠经验、靠催、靠补锅把项目往前推。

我想做的是一个更像”软件工程项目经理助手”的系统：把一段模糊的想法拆成模块、功能点、任务、测试用例、验收标准和风险提醒；让 AI 先把脏活、细活、重复活做起来，再由人来判断、修正、确认。它不应该替代项目经理、产品、研发或测试，而应该让大家少一点互相猜，多一点共同看见。

如果你也经历过这些场景，可能会懂这个项目想解决的东西：

- 需求说了很多，但没人能立刻拆成可执行任务。
- 开发做完了，才发现测试点和验收标准没跟上。
- 项目进度看起来还行，但其实风险、缺陷、变更已经埋在下面。
- AI 能写很多内容，但缺少一个工程化的地方把它变成项目资产。

这个仓库现在只是一个起点。欢迎有缘人一起把它做下去：你可以写后端、做前端、设计交互、补测试、研究 AI prompt、完善项目管理流程，或者只是提出一个你在真实项目里踩过的坑。只要这个系统能让项目从”靠人硬扛”变成”有结构地推进”，它就值得继续长大。

## What It Does

这个项目目标是做一个自建 Web 管理后台，帮助负责人、项目经理/产品、研发、测试、验收人围绕同一套项目数据完成：

- ✅ 原始需求录入
- ✅ AI 分阶段拆分需求
- ✅ 开发任务生成
- ✅ 测试用例生成
- ✅ AI 辅助测试
- ✅ 人工复核测试结论
- ✅ 验收检查项管理
- ✅ 缺陷流转
- ✅ 需求变更影响分析
- ✅ 项目健康度感知

第一版后端使用 Go 实现，前端使用 Vue 3 + Element Plus，AI 能力通过可替换的 Provider 接口实现，当前支持 stub provider（用于测试）和 OpenAI compatible provider。

## Current Status

✅ **已完成功能**：

### 后端 API
- 项目 CRUD 基础接口
- 原始需求录入接口
- AI 草稿生成接口（支持 stub 和 OpenAI provider）
- AI 草稿结构化校验
- AI 草稿保存到 PostgreSQL
- **AI 草稿发布为正式计划**
- **任务看板接口**
- **测试用例确认与 AI 测试执行**
- **缺陷流转接口（含自动回归）**
- **需求变更流程（含 AI 影响分析）**
- **健康度计算（规则引擎 + AI 洞察）**
- 完整的单元测试覆盖

### 前端界面
- 项目列表页
- 项目总览页
- 需求拆分页
- 任务看板页
- 测试验收页
- 缺陷管理页
- 需求变更页
- 风险与缺口页

### 核心特性
- **AI 输出先进草稿区**：人工确认后才发布为正式数据
- **AI 测试结果需人工复核**：不直接作为最终结论
- **模型无关架构**：业务层不依赖具体 AI 供应商
- **结构化输出校验**：确保 AI 输出符合业务规则
- **自动化回归测试**：缺陷修复后自动触发 AI 重测
- **变更影响分析**：AI 分析需求变更对现有计划的影响

## Tech Stack

- **后端**: Go 1.25+, Chi HTTP router, PostgreSQL 16, pgx
- **前端**: Vite + Vue 3 + Element Plus
- **AI**: 可替换 Provider 接口（Stub / OpenAI Compatible）
- **部署**: Docker Compose
- **测试**: Go test, 端到端测试脚本

## Project Structure

```text
cmd/
  api/          API server entrypoint
  migrate/      simple migration runner
internal/
  ai/           AI task interface, stub provider, OpenAI provider
  config/       environment config
  domain/       models and enums
  http/         router and handlers
  repository/   PostgreSQL repositories
  service/      business services
migrations/     database migrations
frontend/       Vue 3 + Element Plus frontend
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
{“status”:”ok”}
```

### 4. Start Frontend (Optional)

```bash
cd frontend
npm install
npm run dev
```

Frontend will be available at:

```text
http://localhost:5173
```

## Environment Variables

See `.env.example`.

```bash
APP_ADDR=:8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/reqsplit?sslmode=disable

# AI Provider 配置
AI_PROVIDER=stub                    # 可选: stub, openai
AI_API_KEY=                         # OpenAI API Key (使用 openai provider 时需要)
AI_API_URL=https://api.openai.com  # OpenAI API URL (可选，默认官方地址)
AI_MODEL=gpt-4                      # 模型名称 (可选，默认 gpt-4)
```

## API Quick Test

完整的端到端测试流程：

```bash
chmod +x test-e2e.sh
./test-e2e.sh
```

或手动测试：

### 1. Create a project

```bash
curl -s -X POST http://localhost:8080/api/projects \
  -H 'Content-Type: application/json' \
  -d '{
    “name”: “AI项目经理系统”,
    “objective”: “自动拆分需求并跟踪验收”,
    “scope”: “第一版闭环”
  }'
```

### 2. Add raw requirement

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/requirements \
  -H 'Content-Type: application/json' \
  -d '{
    “title”: “用户登录功能”,
    “content”: “我要做一个项目需求自动拆分的系统，AI 写测试用例，也能辅助测试，人工复核。”
  }'
```

### 3. Generate AI split draft

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/ai/split-requirement \
  -H 'Content-Type: application/json' \
  -d '{
    “requirement_id”: “<requirement_id>”,
    “content”: “我要做一个项目需求自动拆分的系统，AI 写测试用例，也能辅助测试，人工复核。”
  }'
```

### 4. Publish draft to formal plan

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/ai-drafts/<draft_id>/publish
```

### 5. Get formal plan

```bash
curl -s http://localhost:8080/api/projects/<project_id>/plan
```

### 6. Update task status

```bash
curl -s -X PATCH http://localhost:8080/api/projects/<project_id>/dev-tasks/<task_id>/status \
  -H 'Content-Type: application/json' \
  -d '{“status”: “developing”}'
```

### 7. Confirm test case

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/test-cases/<test_case_id>/confirm
```

### 8. Run AI test

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/test-cases/<test_case_id>/ai-run
```

### 9. Review test result

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/test-runs/<test_run_id>/review \
  -H 'Content-Type: application/json' \
  -d '{“status”: “passed”}'
```

### 10. Create defect

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/defects \
  -H 'Content-Type: application/json' \
  -d '{
    “title”: “登录按钮无响应”,
    “description”: “点击登录按钮后没有任何反应”
  }'
```

### 11. Update defect status (triggers auto regression)

```bash
curl -s -X PATCH http://localhost:8080/api/projects/<project_id>/defects/<defect_id>/status \
  -H 'Content-Type: application/json' \
  -d '{“status”: “pending_regression”}'
```

### 12. Create change request

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/changes \
  -H 'Content-Type: application/json' \
  -d '{
    “title”: “增加手机号登录”,
    “content”: “需要支持手机号+验证码登录方式”
  }'
```

### 13. Analyze change impact

```bash
curl -s -X POST http://localhost:8080/api/projects/<project_id>/changes/<change_id>/analyze
```

### 14. Get project health

```bash
curl -s http://localhost:8080/api/projects/<project_id>/health
```

## Tests

Run all tests:

```bash
go test ./...
```

Run service tests with verbose output:

```bash
go test ./internal/service -v
```

Run end-to-end test:

```bash
./test-e2e.sh
```

## API Documentation

详细的 API 文档请参考：[docs/api.md](docs/api.md)

## Design Docs

- 产品设计：`docs/superpowers/specs/2026-05-15-project-requirement-splitting-design.md`
- 实现计划：`docs/superpowers/plans/2026-05-15-project-requirement-splitting-go-implementation.md`

## Key Features

### 1. AI 草稿机制
- AI 生成的内容先进入草稿区
- 人工审核、编辑后才能发布为正式计划
- 保证 AI 输出的可控性和可追溯性

### 2. 测试用例人工确认
- AI 生成的测试用例需要测试人员确认
- 未确认的测试用例不计入正式测试计划
- 确保测试覆盖率的准确性

### 3. AI 测试结果复核
- AI 执行测试后生成测试结果和证据
- 测试人员复核后才能作为最终结论
- 支持通过、失败、需重测、忽略四种复核状态

### 4. 自动化回归测试
- 缺陷状态变更为”待回归”时自动触发 AI 重测
- 异步执行，不阻塞主流程
- 提高回归测试效率

### 5. 需求变更影响分析
- AI 分析变更对模块、功能点、任务、测试用例的影响
- 生成轻量级计划摘要，避免 Token 超限
- 辅助决策变更是否接受

### 6. 项目健康度
- 规则引擎计算基础健康分（0-100）
- AI 生成管理洞察和建议
- 综合考虑测试覆盖率、活跃缺陷、任务进度等指标

## Architecture Highlights

### 模型无关设计
业务层通过 `ai.Provider` 接口调用 AI 能力，不依赖具体供应商：

```go
type Provider interface {
    Run(ctx context.Context, input TaskInput) (TaskOutput, error)
}
```

当前支持：
- **StubProvider**: 返回确定性测试数据，用于开发和测试
- **OpenAIProvider**: 兼容 OpenAI API 的真实模型调用

### 分层架构
- **Domain**: 领域模型和枚举
- **Repository**: 数据持久化接口和实现
- **Service**: 业务逻辑和工作流
- **HTTP**: 路由和 Handler
- **AI**: AI 任务接口和 Provider 实现

### 数据流
```
原始需求 → AI 拆分草稿 → 人工确认 → 正式计划 
→ 开发任务 → AI 生成测试用例 → 人工确认测试用例 
→ AI 执行测试 → 人工复核 → 缺陷流转 → 自动回归
```

## Roadmap

未来可以扩展：

- [ ] 跨项目资源排期
- [ ] 多模型任务路由
- [ ] 第三方系统导入导出（Jira、禅道等）
- [ ] 更复杂的组织权限
- [ ] 项目经营分析
- [ ] 自动生成周报和验收报告
- [ ] 基于历史项目的拆分质量评估
- [ ] 实时协作和通知
- [ ] 移动端支持

## Contributing

欢迎贡献代码、提出建议或报告问题！

## License

MIT
