package domain

import (
	"encoding/json"
	"time"
)

// LLMTaskSummary is a compact representation of a task sent to the LLM.
type LLMTaskSummary struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	Tags        []string `json:"tags,omitempty"`
	DueDate     *string  `json:"due_date,omitempty"` // RFC3339
	WaitingDays int      `json:"waiting_days,omitempty"`
	BlocksCount int      `json:"blocks_count,omitempty"`
	SourceType  string   `json:"source_type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// LLMPrioritizationRequest is the full payload sent to the LLM for batch prioritization.
type LLMPrioritizationRequest struct {
	Tasks       []LLMTaskSummary `json:"tasks"`
	UserContext string           `json:"user_context,omitempty"`
	MaxResults  int              `json:"max_results"`
}

// LLMPrioritizationResult holds the LLM's ordered list of task IDs.
type LLMPrioritizationResult struct {
	OrderedTaskIDs []string          `json:"ordered_task_ids"`
	Reasoning      map[string]string `json:"reasoning,omitempty"`
}

// LLMRequest is the normalized request sent to any LLM provider.
type LLMRequest struct {
	SystemPrompt string
	UserPrompt   string
	Model        string
	MaxTokens    int
}

// LLMResponse is the raw response from the LLM API.
type LLMResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
}

// BuildPrioritizationPrompt constructs the LLM prompt from a prioritization request.
func BuildPrioritizationPrompt(req LLMPrioritizationRequest) (LLMRequest, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return LLMRequest{}, err
	}

	system := `You are a productivity assistant helping to prioritize tasks.
Given a list of tasks with metadata, return a JSON object with:
- "ordered_task_ids": array of task IDs ordered from highest to lowest priority
- "reasoning": object mapping task IDs to one-sentence explanations

Consider: deadlines, blocking relationships, business impact, waiting time.
Return only valid JSON with no additional text.`

	return LLMRequest{
		SystemPrompt: system,
		UserPrompt:   string(payload),
		MaxTokens:    4096,
	}, nil
}

// TaskToLLMSummary converts a Task to its LLM-friendly representation.
func TaskToLLMSummary(t Task) LLMTaskSummary {
	s := LLMTaskSummary{
		ID:          string(t.ID),
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		Tags:        t.Tags,
		BlocksCount: t.Priority.BlocksCount,
		WaitingDays: t.Priority.WaitingDays,
		CreatedAt:   t.CreatedAt,
	}
	if t.DueDate != nil {
		due := t.DueDate.Format(time.RFC3339)
		s.DueDate = &due
	}
	if len(t.SourceItems) > 0 {
		s.SourceType = string(t.SourceItems[0].SourceType)
	}
	return s
}
