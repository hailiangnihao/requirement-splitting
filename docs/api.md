# API 接口清单

默认地址：

```text
http://localhost:8080
```

所有接口返回 JSON。服务已开放 CORS，支持浏览器跨域调用和 `OPTIONS` 预检。

## 健康检查

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/healthz` | 服务健康检查 |

## 项目

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/projects` | 创建项目 |
| GET | `/api/projects` | 项目列表 |
| GET | `/api/projects/{id}` | 项目详情 |

创建项目请求体：

```json
{
  "name": "AI项目经理系统",
  "objective": "自动拆分需求并跟踪验收",
  "scope": "第一版闭环",
  "owner_id": "",
  "owner_role": "owner"
}
```

## 原始需求

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/projects/{id}/requirements` | 新增原始需求 |
| GET | `/api/projects/{id}/requirements` | 原始需求列表 |

新增原始需求请求体：

```json
{
  "title": "原始需求",
  "content": "我要做一个项目需求自动拆分的系统。",
  "source_type": "manual",
  "source_filename": "",
  "created_by": ""
}
```

## AI 草稿

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/projects/{id}/ai/split-requirement` | AI 拆分原始需求 |
| GET | `/api/projects/{id}/ai-drafts` | AI 草稿列表 |
| POST | `/api/projects/{project_id}/ai-drafts/{draft_id}/publish` | 发布草稿为正式计划 |

AI 拆分请求体：

```json
{
  "requirement_id": "requirement-id",
  "content": "需求正文",
  "created_by": ""
}
```

## 正式计划

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/projects/{project_id}/plan` | 获取模块、功能点、任务、测试用例、验收项树 |
| GET | `/api/projects/{project_id}/dev-tasks` | 开发任务列表 |
| PATCH | `/api/projects/{project_id}/dev-tasks/{id}/status` | 更新开发任务状态 |
| GET | `/api/projects/{project_id}/test-cases` | 测试用例列表 |

更新开发任务状态请求体：

```json
{
  "status": "developing"
}
```

任务状态枚举：

```text
pending_dev
developing
pending_test
testing
pending_acceptance
accepted
launched
```

## 测试执行

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/projects/{project_id}/test-cases/{id}/confirm` | 人工确认测试用例 |
| POST | `/api/projects/{project_id}/test-cases/{id}/ai-run` | 触发 AI 测试执行 |
| GET | `/api/projects/{project_id}/test-runs` | 测试执行记录列表 |
| POST | `/api/projects/{project_id}/test-runs/{id}/review` | 人工复核测试执行 |

人工复核请求体：

```json
{
  "status": "passed"
}
```

复核状态枚举：

```text
passed
failed
needs_retest
ignored
```

## 缺陷

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/projects/{project_id}/defects` | 创建缺陷 |
| GET | `/api/projects/{project_id}/defects` | 缺陷列表 |
| PATCH | `/api/projects/{project_id}/defects/{id}/status` | 更新缺陷状态 |

创建缺陷请求体：

```json
{
  "title": "登录后头像未展示",
  "description": "AI 测试发现头像资源缺失。",
  "test_run_id": "test-run-id",
  "created_by": ""
}
```

更新缺陷状态请求体：

```json
{
  "status": "pending_regression"
}
```

缺陷状态枚举：

```text
pending_confirm
pending_fix
fixing
pending_regression
regression_passed
closed
rejected
```

## 需求变更

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/projects/{project_id}/changes` | 提交需求变更 |
| GET | `/api/projects/{project_id}/changes` | 需求变更列表 |
| POST | `/api/projects/{project_id}/changes/{id}/analyze` | AI 分析变更影响 |
| PATCH | `/api/projects/{project_id}/changes/{id}/status` | 更新变更状态 |

提交变更请求体：

```json
{
  "title": "增加微信扫码登录",
  "content": "登录方式增加微信扫码。",
  "created_by": ""
}
```

更新变更状态请求体：

```json
{
  "status": "accepted"
}
```

变更状态枚举：

```text
submitted
analyzed
accepted
applied
rejected
```

## 健康度

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/projects/{project_id}/health` | 项目健康度指标和 AI 洞察 |
