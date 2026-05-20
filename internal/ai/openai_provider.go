package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAIProvider struct {
	apiKey string
	apiURL string
	model  string
	client *http.Client
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
		client: &http.Client{},
	}
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
			continue // 网络错误，触发重试
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close() // 循环体内显式关闭，避免 defer 导致的资源泄漏

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(body))
			continue // 接口报错 (例如 429 限流, 502 等)，触发重试
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

// getSystemPrompt 为不同的任务分配系统级 Prompt（注意：开启 JSON 模式，提示词中必须包含 "JSON" 字眼）
func (p *OpenAIProvider) getSystemPrompt(taskType TaskType) string {
	switch taskType {
	case TaskSplitRequirement:
		return "你是一个资深敏捷项目经理。请分析用户提供的原始需求，并将其拆分为模块(modules)、里程碑(milestones)、功能点(feature_points)、开发任务(dev_tasks)、测试用例(test_cases)和验收项(acceptance_items)。请必须以 JSON 格式返回，确保层级结构清晰。"
	case TaskType("execute_ai_test"):
		return "你是一个高级自动化测试工程师。根据用户提供的测试用例上下文，模拟执行测试并生成结果。必须返回合法的 JSON 对象，包含 actual_result (字符串), is_defect_suggested (布尔值), evidence (对象，可包含 screenshots, logs 等键)。"
	case TaskAnalyzeChangeImpact:
		return "你是一个资深架构师兼测试总监。请根据当前的项目计划树和变更诉求，分析变更影响面。必须返回合法的 JSON 对象，包含 risk_level, summary, affected_feature_points (字符串数组), affected_test_cases (对象数组), new_tasks_suggested (对象数组), estimated_extra_days。"
	case TaskType("generate_health_insight"):
		return "你是一个拥有 20 年经验的资深敏捷技术总监。请根据以下项目的实时运行指标，指出当前项目最大的 2 个隐患，并给技术团队提出具体的行动建议。必须返回 JSON，包含 health_status (字符串, 例如 at_risk, healthy), executive_summary (字符串), top_risks (对象数组, 包含 title, description), action_items (字符串数组)。"
	default:
		return ""
	}
}
