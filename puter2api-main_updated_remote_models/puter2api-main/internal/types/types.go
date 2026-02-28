package types

import "encoding/json"

// ==================== Claude API 类型 ====================

// ClaudeRequest Claude API 请求结构
type ClaudeRequest struct {
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	Stream    bool            `json:"stream"`
	Tools     json.RawMessage `json:"tools,omitempty"`
	System    json.RawMessage `json:"system,omitempty"`
}

// ClaudeMessage Claude 消息
type ClaudeMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// ContentBlock 通用内容块（支持 text, tool_use, tool_result）
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

// TextContentBlock 文本内容块（text 字段始终存在）
type TextContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolUseContentBlock 工具使用内容块
type ToolUseContentBlock struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// Usage token 使用量
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// DeltaUsage 增量使用量
type DeltaUsage struct {
	OutputTokens int `json:"output_tokens"`
}

// ToolDef 工具定义
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

// ParsedToolCall 解析后的工具调用
type ParsedToolCall struct {
	Name  string          `json:"name"`
	ID    string          `json:"id,omitempty"`
	Input json.RawMessage `json:"input"`
}

// ==================== Claude SSE 事件类型 ====================

// MessageStartEvent message_start 事件
type MessageStartEvent struct {
	Type    string             `json:"type"`
	Message MessageStartDetail `json:"message"`
}

// MessageStartDetail message_start 详情
type MessageStartDetail struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   *string        `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence"`
	Usage        Usage          `json:"usage"`
}

// ContentBlockStartEvent content_block_start 事件
type ContentBlockStartEvent struct {
	Type         string      `json:"type"`
	Index        int         `json:"index"`
	ContentBlock interface{} `json:"content_block"`
}

// ContentBlockDeltaEvent content_block_delta 事件
type ContentBlockDeltaEvent struct {
	Type  string      `json:"type"`
	Index int         `json:"index"`
	Delta interface{} `json:"delta"`
}

// TextDelta 文本增量
type TextDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// InputJSONDelta JSON 输入增量
type InputJSONDelta struct {
	Type        string `json:"type"`
	PartialJSON string `json:"partial_json"`
}

// ContentBlockStopEvent content_block_stop 事件
type ContentBlockStopEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// MessageDeltaEvent message_delta 事件
type MessageDeltaEvent struct {
	Type  string       `json:"type"`
	Delta MessageDelta `json:"delta"`
	Usage DeltaUsage   `json:"usage"`
}

// MessageDelta 消息增量
type MessageDelta struct {
	StopReason   string  `json:"stop_reason"`
	StopSequence *string `json:"stop_sequence"`
}

// MessageStopEvent message_stop 事件
type MessageStopEvent struct {
	Type string `json:"type"`
}

// ==================== Puter API 类型 ====================

// PuterRequest Puter API 请求
type PuterRequest struct {
	Interface string    `json:"interface"`
	Driver    string    `json:"driver"`
	TestMode  bool      `json:"test_mode"`
	Method    string    `json:"method"`
	Args      PuterArgs `json:"args"`
	AuthToken string    `json:"auth_token"`
}

// PuterArgs Puter 请求参数
type PuterArgs struct {
	Messages []PuterMessage `json:"messages"`
	Model    string         `json:"model"`
	Stream   bool           `json:"stream"`
}

// PuterMessage Puter 消息
type PuterMessage struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content"`
}

// PuterStreamChunk Puter 流式响应块
type PuterStreamChunk struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ==================== OpenAI API 类型 ====================

// OpenAIRequest OpenAI Chat Completion 请求
type OpenAIRequest struct {
	Model            string          `json:"model"`
	Messages         []OpenAIMessage `json:"messages"`
	MaxTokens        int             `json:"max_tokens,omitempty"`
	Temperature      float64         `json:"temperature,omitempty"`
	TopP             float64         `json:"top_p,omitempty"`
	N                int             `json:"n,omitempty"`
	Stream           bool            `json:"stream,omitempty"`
	Stop             json.RawMessage `json:"stop,omitempty"`
	PresencePenalty  float64         `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
	Tools            []OpenAITool    `json:"tools,omitempty"`
	ToolChoice       json.RawMessage `json:"tool_choice,omitempty"`
}

// OpenAIMessage OpenAI 消息
type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    json.RawMessage  `json:"content"`
	Name       string           `json:"name,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

// OpenAITool OpenAI 工具定义
type OpenAITool struct {
	Type     string             `json:"type"`
	Function OpenAIToolFunction `json:"function"`
}

// OpenAIToolFunction OpenAI 工具函数定义
type OpenAIToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// OpenAIToolCall OpenAI 工具调用
type OpenAIToolCall struct {
	Index    *int                   `json:"index,omitempty"`
	ID       string                 `json:"id,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Function OpenAIToolCallFunction `json:"function"`
}

// OpenAIToolCallFunction OpenAI 工具调用函数
type OpenAIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// OpenAIResponse OpenAI Chat Completion 响应
type OpenAIResponse struct {
	ID                string         `json:"id"`
	Object            string         `json:"object"`
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	Choices           []OpenAIChoice `json:"choices"`
	Usage             *OpenAIUsage   `json:"usage,omitempty"`
	SystemFingerprint string         `json:"system_fingerprint,omitempty"`
}

// OpenAIChoice OpenAI 选择
type OpenAIChoice struct {
	Index        int                `json:"index"`
	Message      *OpenAIResponseMsg `json:"message,omitempty"`
	Delta        *OpenAIResponseMsg `json:"delta,omitempty"`
	FinishReason *string            `json:"finish_reason"`
	Logprobs     *interface{}       `json:"logprobs"`
}

// OpenAIResponseMsg OpenAI 响应消息
type OpenAIResponseMsg struct {
	Role      string           `json:"role,omitempty"`
	Content   *string          `json:"content"`
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
}

// OpenAIUsage OpenAI 使用量
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
