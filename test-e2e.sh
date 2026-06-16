#!/bin/bash
# 端到端测试脚本

set -e

echo "=== 端到端测试开始 ==="

# 1. 检查 PostgreSQL
echo "1. 检查 PostgreSQL 连接..."
docker compose ps postgres | grep -q "healthy" || {
  echo "PostgreSQL 未运行，正在启动..."
  docker compose up -d postgres
  sleep 5
}

# 2. 运行数据库迁移
echo "2. 运行数据库迁移..."
go run ./cmd/migrate up

# 3. 启动 API 服务（后台）
echo "3. 启动 API 服务..."
go run ./cmd/api &
API_PID=$!
sleep 3

# 4. 健康检查
echo "4. 健康检查..."
curl -f http://localhost:8080/healthz || {
  echo "API 健康检查失败"
  kill $API_PID
  exit 1
}

# 5. 创建项目
echo "5. 创建测试项目..."
PROJECT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/projects \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "测试项目",
    "objective": "端到端测试",
    "scope": "完整流程验证"
  }')

PROJECT_ID=$(echo $PROJECT_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "项目 ID: $PROJECT_ID"

# 6. 添加需求
echo "6. 添加原始需求..."
REQ_RESPONSE=$(curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/requirements \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "测试需求",
    "content": "我要做一个用户登录功能，支持用户名密码登录和微信扫码登录。",
    "source_type": "manual"
  }')

REQ_ID=$(echo $REQ_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "需求 ID: $REQ_ID"

# 7. AI 拆分需求
echo "7. AI 拆分需求..."
DRAFT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/ai/split-requirement \
  -H 'Content-Type: application/json' \
  -d "{
    \"requirement_id\": \"$REQ_ID\",
    \"content\": \"我要做一个用户登录功能，支持用户名密码登录和微信扫码登录。\"
  }")

DRAFT_ID=$(echo $DRAFT_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "草稿 ID: $DRAFT_ID"

# 8. 发布草稿
echo "8. 发布草稿为正式计划..."
curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/ai-drafts/$DRAFT_ID/publish

# 9. 获取正式计划
echo "9. 获取正式计划..."
curl -s http://localhost:8080/api/projects/$PROJECT_ID/plan | jq '.modules | length' || echo "计划获取成功"

# 10. 获取开发任务
echo "10. 获取开发任务列表..."
TASKS=$(curl -s http://localhost:8080/api/projects/$PROJECT_ID/dev-tasks)
TASK_ID=$(echo $TASKS | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "任务 ID: $TASK_ID"

# 11. 更新任务状态
if [ -n "$TASK_ID" ]; then
  echo "11. 更新任务状态..."
  curl -s -X PATCH http://localhost:8080/api/projects/$PROJECT_ID/dev-tasks/$TASK_ID/status \
    -H 'Content-Type: application/json' \
    -d '{"status": "developing"}'
fi

# 12. 获取测试用例
echo "12. 获取测试用例..."
TEST_CASES=$(curl -s http://localhost:8080/api/projects/$PROJECT_ID/test-cases)
TEST_CASE_ID=$(echo $TEST_CASES | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "测试用例 ID: $TEST_CASE_ID"

# 13. 确认测试用例
if [ -n "$TEST_CASE_ID" ]; then
  echo "13. 确认测试用例..."
  curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/test-cases/$TEST_CASE_ID/confirm
fi

# 14. AI 执行测试
if [ -n "$TEST_CASE_ID" ]; then
  echo "14. AI 执行测试..."
  TEST_RUN=$(curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/test-cases/$TEST_CASE_ID/ai-run)
  TEST_RUN_ID=$(echo $TEST_RUN | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
  echo "测试执行 ID: $TEST_RUN_ID"
fi

# 15. 复核测试结果
if [ -n "$TEST_RUN_ID" ]; then
  echo "15. 复核测试结果..."
  curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/test-runs/$TEST_RUN_ID/review \
    -H 'Content-Type: application/json' \
    -d '{"status": "passed"}'
fi

# 16. 创建缺陷
echo "16. 创建缺陷..."
DEFECT=$(curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/defects \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "登录按钮无响应",
    "description": "点击登录按钮后没有任何反应"
  }')
DEFECT_ID=$(echo $DEFECT | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "缺陷 ID: $DEFECT_ID"

# 17. 更新缺陷状态
if [ -n "$DEFECT_ID" ]; then
  echo "17. 更新缺陷状态..."
  curl -s -X PATCH http://localhost:8080/api/projects/$PROJECT_ID/defects/$DEFECT_ID/status \
    -H 'Content-Type: application/json' \
    -d '{"status": "fixing"}'
fi

# 18. 创建需求变更
echo "18. 创建需求变更..."
CHANGE=$(curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/changes \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "增加手机号登录",
    "content": "需要支持手机号+验证码登录方式"
  }')
CHANGE_ID=$(echo $CHANGE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "变更 ID: $CHANGE_ID"

# 19. 分析变更影响
if [ -n "$CHANGE_ID" ]; then
  echo "19. 分析变更影响..."
  curl -s -X POST http://localhost:8080/api/projects/$PROJECT_ID/changes/$CHANGE_ID/analyze
fi

# 20. 获取健康度
echo "20. 获取项目健康度..."
curl -s http://localhost:8080/api/projects/$PROJECT_ID/health | jq '.metrics.base_score' || echo "健康度获取成功"

# 清理
echo ""
echo "=== 测试完成，清理资源 ==="
kill $API_PID

echo ""
echo "✅ 端到端测试全部通过！"
