package ai

import "context"

type TaskType string

const (
	TaskSplitRequirement    TaskType = "split_requirement"
	TaskGenerateTestCases   TaskType = "generate_test_cases"
	TaskExecuteAITest       TaskType = "execute_ai_test"
	TaskAnalyzeChangeImpact TaskType = "analyze_change_impact"
	TaskExplainRisk         TaskType = "explain_risk"
)

type TaskInput struct {
	Type      TaskType       `json:"type"`
	ProjectID string         `json:"project_id"`
	Payload   map[string]any `json:"payload"`
}

type TaskOutput struct {
	Type   TaskType       `json:"type"`
	Result map[string]any `json:"result"`
}

type Provider interface {
	Run(ctx context.Context, input TaskInput) (TaskOutput, error)
}

type SplitRequirementResult struct {
	Modules         []ModuleDraft         `json:"modules"`
	Milestones      []MilestoneDraft      `json:"milestones"`
	FeaturePoints   []FeaturePointDraft   `json:"feature_points"`
	DevTasks        []DevTaskDraft        `json:"dev_tasks"`
	TestCases       []TestCaseDraft       `json:"test_cases"`
	AcceptanceItems []AcceptanceItemDraft `json:"acceptance_items"`
	Risks           []RiskDraft           `json:"risks"`
}

type ModuleDraft struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MilestoneDraft struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type FeaturePointDraft struct {
	Key          string `json:"key"`
	ModuleKey    string `json:"module_key"`
	MilestoneKey string `json:"milestone_key"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Priority     string `json:"priority"`
}

type DevTaskDraft struct {
	Key             string `json:"key"`
	FeaturePointKey string `json:"feature_point_key"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Priority        string `json:"priority"`
}

type TestCaseDraft struct {
	Key             string   `json:"key"`
	FeaturePointKey string   `json:"feature_point_key"`
	DevTaskKey      string   `json:"dev_task_key"`
	Title           string   `json:"title"`
	CaseType        string   `json:"case_type"`
	Priority        string   `json:"priority"`
	Preconditions   string   `json:"preconditions"`
	Steps           []string `json:"steps"`
	TestData        string   `json:"test_data"`
	ExpectedResult  string   `json:"expected_result"`
}

type AcceptanceItemDraft struct {
	Key             string `json:"key"`
	FeaturePointKey string `json:"feature_point_key"`
	MilestoneKey    string `json:"milestone_key"`
	Title           string `json:"title"`
	PassCriteria    string `json:"pass_criteria"`
}

type RiskDraft struct {
	Key         string `json:"key"`
	Category    string `json:"category"`
	Level       string `json:"level"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
