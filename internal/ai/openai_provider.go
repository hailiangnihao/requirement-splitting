package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type OpenAIProvider struct {
	apiKey      string
	apiURL      string
	model       string
	client      *http.Client
	rateMutex   sync.Mutex
	lastCall    time.Time
	minInterval time.Duration // 最小请求间隔
}

// StreamProgress 流式进度消息
type StreamProgress struct {
	Type       string      `json:"type"`        // "progress", "thinking", "content", "result", "error"
	Message    string      `json:"message"`     // 进度消息
	Thinking   string      `json:"thinking"`    // 思考过程
	Content    string      `json:"content"`     // 生成的内容
	Data       interface{} `json:"data"`        // 结果数据
	Progress   int         `json:"progress"`    // 进度百分比
}

func NewOpenAIProvider(apiKey, apiURL, model string) *OpenAIProvider {
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{
		apiKey: apiKey,
		apiURL: apiURL,
		model:  model,
		client: &http.Client{
			Timeout: 60 * time.Second, // 设置 60 秒超时
		},
		minInterval: 2 * time.Second, // 默认每次请求至少间隔 2 秒
	}
}

// waitForRateLimit 等待满足速率限制
func (p *OpenAIProvider) waitForRateLimit() {
	p.rateMutex.Lock()
	defer p.rateMutex.Unlock()

	elapsed := time.Since(p.lastCall)
	if elapsed < p.minInterval {
		time.Sleep(p.minInterval - elapsed)
	}
	p.lastCall = time.Now()
}

func (p *OpenAIProvider) Run(ctx context.Context, input TaskInput) (TaskOutput, error) {
	// 设定最大重试次数
	const maxRetries = 3
	var lastErr error

	systemPrompt := p.getSystemPrompt(input.Type)
	if systemPrompt == "" {
		return TaskOutput{}, fmt.Errorf("openai provider does not support task type: %s", input.Type)
	}

	payloadBytes, _ := json.Marshal(input.Payload)
	userPrompt := fmt.Sprintf("请根据以下提供的数据上下文完成任务：\n%s", string(payloadBytes))

	reqBody := map[string]any{
		"model":           p.model,
		"response_format": map[string]string{"type": "json_object"}, // 强制返回 JSON
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	reqBytes, _ := json.Marshal(reqBody)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 等待满足速率限制
		p.waitForRateLimit()

		// 注意：每次请求都需要重新创建 Reader
		req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, bytes.NewReader(reqBytes))
		if err != nil {
			return TaskOutput{}, err // 构建请求失败直接退出，无需重试
		}

		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to call openai: %w", err)
			// 网络错误，指数退避后重试
			if attempt < maxRetries {
				waitTime := time.Duration(1<<uint(attempt-1)) * time.Second
				time.Sleep(waitTime)
			}
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close() // 循环体内显式关闭，避免 defer 导致的资源泄漏

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(body))

			// 针对 429 限流，使用更长的退避时间
			if resp.StatusCode == http.StatusTooManyRequests {
				if attempt < maxRetries {
					// 429 错误：等待 5s, 10s, 20s
					waitTime := time.Duration(5*(1<<uint(attempt-1))) * time.Second
					time.Sleep(waitTime)
				}
				continue
			}

			// 其他错误（如 502, 503），指数退避
			if attempt < maxRetries {
				waitTime := time.Duration(1<<uint(attempt-1)) * time.Second
				time.Sleep(waitTime)
			}
			continue
		}

		var openaiResp struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(body, &openaiResp); err != nil {
			lastErr = fmt.Errorf("failed to decode openai response: %w", err)
			continue // 响应体格式异常，触发重试
		}

		rawContent := openaiResp.Choices[0].Message.Content
		var result map[string]any
		if err := json.Unmarshal([]byte(rawContent), &result); err != nil {
			// 核心防御：JSON 反序列化失败，通常是因为大模型幻觉输出了非 JSON 字符
			// 我们把 rawContent 塞进 lastErr 里，这样如果 3 次都失败了，日志里能看到大模型到底吐了什么垃圾数据
			lastErr = fmt.Errorf("failed to parse result as json object: %w\nRaw Content: %s", err, rawContent)
			continue // 格式损毁，触发重试
		}

		// 成功解析 JSON，直接返回
		return TaskOutput{Type: input.Type, Result: result}, nil
	}

	// 如果跑完了循环还没 return，说明全军覆没
	return TaskOutput{}, fmt.Errorf("task failed after %d attempts, last error: %w", maxRetries, lastErr)
}

// RunStream 流式执行任务，通过回调函数实时返回进度
func (p *OpenAIProvider) RunStream(ctx context.Context, input TaskInput, progressCallback func(StreamProgress)) (TaskOutput, error) {
	// 设定最大重试次数
	const maxRetries = 3
	var lastErr error

	systemPrompt := p.getSystemPrompt(input.Type)
	if systemPrompt == "" {
		return TaskOutput{}, fmt.Errorf("openai provider does not support task type: %s", input.Type)
	}

	payloadBytes, _ := json.Marshal(input.Payload)
	userPrompt := fmt.Sprintf("请根据以下提供的数据上下文完成任务：\n%s", string(payloadBytes))

	reqBody := map[string]any{
		"model":           p.model,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	// 启用流式响应
	reqBody["stream"] = true

	reqBytes, _ := json.Marshal(reqBody)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 等待满足速率限制
		p.waitForRateLimit()

		// 发送进度
		if progressCallback != nil {
			progressCallback(StreamProgress{
				Type:    "progress",
				Message: fmt.Sprintf("正在调用 AI 模型进行拆分... (尝试 %d/%d)", attempt, maxRetries),
			})
		}

		// 注意：每次请求都需要重新创建 Reader
		req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, bytes.NewReader(reqBytes))
		if err != nil {
			return TaskOutput{}, err // 构建请求失败直接退出，无需重试
		}

		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to call openai: %w", err)
			// 网络错误，指数退避后重试
			if attempt < maxRetries {
				waitTime := time.Duration(1<<uint(attempt-1)) * time.Second
				if progressCallback != nil {
					progressCallback(StreamProgress{
						Type:    "progress",
						Message: fmt.Sprintf("网络错误，%v 后重试...", waitTime),
					})
				}
				time.Sleep(waitTime)
			}
			continue
		}

		// 检查是否为流式响应
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			lastErr = fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(body))

			// 针对 429 限流，使用更长的退避时间
			if resp.StatusCode == http.StatusTooManyRequests {
				if attempt < maxRetries {
					waitTime := time.Duration(5*(1<<uint(attempt-1))) * time.Second
					if progressCallback != nil {
						progressCallback(StreamProgress{
							Type:    "progress",
							Message: fmt.Sprintf("触发速率限制，%v 后重试...", waitTime),
						})
					}
					time.Sleep(waitTime)
				}
				continue
			}

			// 其他错误（如 502, 503），指数退避
			if attempt < maxRetries {
				waitTime := time.Duration(1<<uint(attempt-1)) * time.Second
				if progressCallback != nil {
					progressCallback(StreamProgress{
						Type:    "progress",
						Message: fmt.Sprintf("服务错误，%v 后重试...", waitTime),
					})
				}
				time.Sleep(waitTime)
			}
			continue
		}

		// 处理流式响应
		if progressCallback != nil {
			progressCallback(StreamProgress{
				Type:    "progress",
				Message: "AI 正在生成内容...",
			})
		}

		var fullContent strings.Builder
		var fullThinking strings.Builder
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// SSE 格式：data: {...}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// 结束标记
			if data == "[DONE]" {
				break
			}

			// 解析 SSE 数据
			var sseData struct {
				Choices []struct {
					Delta struct {
						Content         string `json:"content"`
						ReasoningContent string `json:"reasoning_content"` // GLM的思考过程
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &sseData); err != nil {
				continue
			}

			if len(sseData.Choices) > 0 {
				// 处理思考过程
				thinking := sseData.Choices[0].Delta.ReasoningContent
				if thinking != "" {
					fullThinking.WriteString(thinking)

					// 实时发送思考过程
					if progressCallback != nil {
						progressCallback(StreamProgress{
							Type:     "thinking",
							Thinking: fullThinking.String(),
							Message:  "AI 正在思考...",
						})
					}
				}

				// 处理生成内容
				content := sseData.Choices[0].Delta.Content
				if content != "" {
					fullContent.WriteString(content)

					// 实时发送内容进度
					if progressCallback != nil {
						progressCallback(StreamProgress{
							Type:     "content",
							Content:  fullContent.String(),
							Message:  fmt.Sprintf("已生成 %d 字节...", fullContent.Len()),
						})
					}
				}
			}
		}
		resp.Body.Close()

		if err := scanner.Err(); err != nil {
			lastErr = fmt.Errorf("failed to read stream: %w", err)
			continue
		}

		// 解析完整响应
		rawContent := fullContent.String()
		var result map[string]any
		if err := json.Unmarshal([]byte(rawContent), &result); err != nil {
			lastErr = fmt.Errorf("failed to parse result as json object: %w\nRaw Content: %s", err, rawContent)
			continue
		}

		// 成功解析 JSON，发送结果
		if progressCallback != nil {
			progressCallback(StreamProgress{
				Type:    "result",
				Message: "AI 拆分完成！",
				Data:    result,
			})
		}

		return TaskOutput{Type: input.Type, Result: result}, nil
	}

	// 如果跑完了循环还没 return，说明全军覆没
	return TaskOutput{}, fmt.Errorf("task failed after %d attempts, last error: %w", maxRetries, lastErr)
}

// getSystemPrompt 为不同的任务分配系统级 Prompt（注意：开启 JSON 模式，提示词中必须包含 "JSON" 字眼）
func (p *OpenAIProvider) getSystemPrompt(taskType TaskType) string {
	switch taskType {
	case TaskSplitRequirement:
		return `你是一个资深敏捷项目经理。请分析用户提供的原始需求，并将其拆分为结构化的项目计划。

**重要要求：**
1. 所有字段名称必须使用英文（如 modules, milestones 等）
2. 所有字段内容（名称、描述等）必须使用中文
3. 必须返回 JSON 格式

**返回结构：**
{
  "modules": [
    {
      "key": "模块唯一标识",
      "name": "模块名称（中文）",
      "description": "模块描述（中文）"
    }
  ],
  "milestones": [
    {
      "key": "里程碑唯一标识",
      "name": "里程碑名称（中文）",
      "description": "里程碑描述（中文）",
      "due_date": "预计完成日期"
    }
  ],
  "feature_points": [
    {
      "key": "功能点唯一标识",
      "module_key": "所属模块key",
      "title": "功能点名称（中文）",
      "description": "功能点描述（中文）"
    }
  ],
  "dev_tasks": [
    {
      "key": "任务唯一标识",
      "feature_point_key": "所属功能点key",
      "title": "任务名称（中文）",
      "description": "任务描述（中文）",
      "priority": "优先级(high/medium/low)"
    }
  ],
  "test_cases": [
    {
      "key": "测试用例唯一标识",
      "feature_point_key": "所属功能点key",
      "title": "测试用例名称（中文）",
      "description": "测试步骤（中文）",
      "expected_result": "预期结果（中文）"
    }
  ],
  "acceptance_items": [
    {
      "key": "验收项唯一标识",
      "feature_point_key": "所属功能点key",
      "title": "验收项名称（中文）",
      "pass_criteria": "通过标准（中文）"
    }
  ]
}`
	case TaskType("execute_ai_test"):
		return `你是一个高级自动化测试工程师。根据用户提供的测试用例上下文，模拟执行测试并生成结果。

**重要要求：**
1. 字段名使用英文
2. 内容使用中文
3. 返回合法的 JSON 对象

**返回结构：**
{
  "actual_result": "实际执行结果（中文）",
  "is_defect_suggested": false,
  "evidence": {
    "screenshots": ["截图说明"],
    "logs": ["日志说明"]
  }
}`
	case TaskAnalyzeChangeImpact:
		return `你是一个资深架构师兼测试总监。请根据当前的项目计划树和变更诉求，分析变更影响面。

**重要要求：**
1. 字段名使用英文
2. 内容使用中文
3. 返回合法的 JSON 对象

**返回结构：**
{
  "risk_level": "风险等级(low/medium/high)",
  "summary": "影响摘要（中文）",
  "affected_feature_points": ["受影响的功能点ID列表"],
  "affected_test_cases": [
    {
      "test_case_id": "测试用例ID",
      "action": "影响操作",
      "reason": "原因（中文）"
    }
  ],
  "new_tasks_suggested": [
    {
      "title": "新任务标题（中文）",
      "description": "任务描述（中文）"
    }
  ],
  "estimated_extra_days": 0
}`
	case TaskType("generate_health_insight"):
		return `你是一个拥有 20 年经验的资深敏捷技术总监。请根据以下项目的实时运行指标，指出当前项目最大的 2 个隐患，并给技术团队提出具体的行动建议。

**重要要求：**
1. 字段名使用英文
2. 内容使用中文
3. 返回 JSON 格式

**返回结构：**
{
  "health_status": "健康状态(healthy/at_risk/critical)",
  "executive_summary": "执行摘要（中文）",
  "top_risks": [
    {
      "title": "风险标题（中文）",
      "description": "风险描述（中文）"
    }
  ],
  "action_items": ["行动建议1（中文）", "行动建议2（中文）"]
}`
	default:
		return ""
	}
}
