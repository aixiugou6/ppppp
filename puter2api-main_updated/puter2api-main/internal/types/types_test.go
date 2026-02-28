package types

import (
	"encoding/json"
	"testing"
)

// ==================== ContentBlock 序列化测试 ====================

func TestContentBlock_TextSerialization(t *testing.T) {
	block := ContentBlock{
		Type: "text",
		Text: "Hello, world!",
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// 验证 JSON 结构
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["type"] != "text" {
		t.Errorf("expected type 'text', got '%v'", result["type"])
	}
	if result["text"] != "Hello, world!" {
		t.Errorf("expected text 'Hello, world!', got '%v'", result["text"])
	}

	// omitempty 字段不应该出现
	if _, exists := result["id"]; exists {
		t.Errorf("id should be omitted when empty")
	}
	if _, exists := result["name"]; exists {
		t.Errorf("name should be omitted when empty")
	}
}

func TestContentBlock_ToolUseSerialization(t *testing.T) {
	block := ContentBlock{
		Type:  "tool_use",
		ID:    "toolu_123",
		Name:  "search",
		Input: json.RawMessage(`{"query": "test"}`),
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["type"] != "tool_use" {
		t.Errorf("expected type 'tool_use'")
	}
	if result["id"] != "toolu_123" {
		t.Errorf("expected id 'toolu_123'")
	}
	if result["name"] != "search" {
		t.Errorf("expected name 'search'")
	}

	input, ok := result["input"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected input to be object")
	}
	if input["query"] != "test" {
		t.Errorf("expected input.query 'test'")
	}
}

func TestContentBlock_ToolResultSerialization(t *testing.T) {
	block := ContentBlock{
		Type:      "tool_result",
		ToolUseID: "toolu_123",
		Content:   json.RawMessage(`"Result content"`),
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["type"] != "tool_result" {
		t.Errorf("expected type 'tool_result'")
	}
	if result["tool_use_id"] != "toolu_123" {
		t.Errorf("expected tool_use_id 'toolu_123'")
	}
}

// ==================== TextContentBlock 测试 ====================

func TestTextContentBlock_AlwaysHasText(t *testing.T) {
	// 即使 text 为空，也应该序列化出 text 字段
	block := TextContentBlock{
		Type: "text",
		Text: "",
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// text 字段必须存在（这是与 ContentBlock 的关键区别）
	if _, exists := result["text"]; !exists {
		t.Errorf("text field should always be present in TextContentBlock")
	}
}

func TestTextContentBlock_WithContent(t *testing.T) {
	block := TextContentBlock{
		Type: "text",
		Text: "Hello!",
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"type":"text","text":"Hello!"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

// ==================== ToolUseContentBlock 测试 ====================

func TestToolUseContentBlock_Serialization(t *testing.T) {
	block := ToolUseContentBlock{
		Type:  "tool_use",
		ID:    "toolu_abc",
		Name:  "calculator",
		Input: json.RawMessage(`{"expression": "2+2"}`),
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["type"] != "tool_use" {
		t.Errorf("expected type 'tool_use'")
	}
	if result["id"] != "toolu_abc" {
		t.Errorf("expected id 'toolu_abc'")
	}
	if result["name"] != "calculator" {
		t.Errorf("expected name 'calculator'")
	}

	// input 字段必须存在
	if _, exists := result["input"]; !exists {
		t.Errorf("input field should always be present")
	}
}

func TestToolUseContentBlock_EmptyInput(t *testing.T) {
	block := ToolUseContentBlock{
		Type:  "tool_use",
		ID:    "toolu_xyz",
		Name:  "no_params_tool",
		Input: json.RawMessage(`{}`),
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	input, ok := result["input"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected input to be object")
	}
	if len(input) != 0 {
		t.Errorf("expected empty input object")
	}
}

// ==================== SSE 事件类型测试 ====================

func TestMessageStartEvent_Serialization(t *testing.T) {
	event := MessageStartEvent{
		Type: "message_start",
		Message: MessageStartDetail{
			ID:      "msg_123",
			Type:    "message",
			Role:    "assistant",
			Content: []ContentBlock{},
			Model:   "claude-3-opus",
			Usage:   Usage{InputTokens: 100, OutputTokens: 0},
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["type"] != "message_start" {
		t.Errorf("expected type 'message_start'")
	}

	msg, ok := result["message"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected message to be object")
	}

	if msg["id"] != "msg_123" {
		t.Errorf("expected message.id 'msg_123'")
	}
	if msg["role"] != "assistant" {
		t.Errorf("expected message.role 'assistant'")
	}
}

func TestContentBlockStartEvent_WithTextBlock(t *testing.T) {
	event := ContentBlockStartEvent{
		Type:         "content_block_start",
		Index:        0,
		ContentBlock: TextContentBlock{Type: "text", Text: ""},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	block, ok := result["content_block"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected content_block to be object")
	}

	if block["type"] != "text" {
		t.Errorf("expected content_block.type 'text'")
	}

	// text 字段必须存在
	if _, exists := block["text"]; !exists {
		t.Errorf("text field should be present in TextContentBlock")
	}
}

func TestContentBlockStartEvent_WithToolUseBlock(t *testing.T) {
	event := ContentBlockStartEvent{
		Type:  "content_block_start",
		Index: 1,
		ContentBlock: ToolUseContentBlock{
			Type:  "tool_use",
			ID:    "toolu_test",
			Name:  "test_tool",
			Input: json.RawMessage(`{}`),
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	block, ok := result["content_block"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected content_block to be object")
	}

	if block["type"] != "tool_use" {
		t.Errorf("expected content_block.type 'tool_use'")
	}
	if block["id"] != "toolu_test" {
		t.Errorf("expected content_block.id 'toolu_test'")
	}
	if block["name"] != "test_tool" {
		t.Errorf("expected content_block.name 'test_tool'")
	}
	if _, exists := block["input"]; !exists {
		t.Errorf("input field should be present in ToolUseContentBlock")
	}
}

func TestContentBlockDeltaEvent_TextDelta(t *testing.T) {
	event := ContentBlockDeltaEvent{
		Type:  "content_block_delta",
		Index: 0,
		Delta: TextDelta{Type: "text_delta", Text: "Hello"},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	delta, ok := result["delta"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected delta to be object")
	}

	if delta["type"] != "text_delta" {
		t.Errorf("expected delta.type 'text_delta'")
	}
	if delta["text"] != "Hello" {
		t.Errorf("expected delta.text 'Hello'")
	}
}

func TestContentBlockDeltaEvent_InputJSONDelta(t *testing.T) {
	event := ContentBlockDeltaEvent{
		Type:  "content_block_delta",
		Index: 1,
		Delta: InputJSONDelta{Type: "input_json_delta", PartialJSON: `{"key":`},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	delta, ok := result["delta"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected delta to be object")
	}

	if delta["type"] != "input_json_delta" {
		t.Errorf("expected delta.type 'input_json_delta'")
	}
	if delta["partial_json"] != `{"key":` {
		t.Errorf("expected delta.partial_json")
	}
}

func TestMessageDeltaEvent_Serialization(t *testing.T) {
	event := MessageDeltaEvent{
		Type:  "message_delta",
		Delta: MessageDelta{StopReason: "end_turn"},
		Usage: DeltaUsage{OutputTokens: 150},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	delta, ok := result["delta"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected delta to be object")
	}

	if delta["stop_reason"] != "end_turn" {
		t.Errorf("expected delta.stop_reason 'end_turn'")
	}

	usage, ok := result["usage"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected usage to be object")
	}

	if usage["output_tokens"] != float64(150) {
		t.Errorf("expected usage.output_tokens 150")
	}
}

func TestMessageStopEvent_Serialization(t *testing.T) {
	event := MessageStopEvent{Type: "message_stop"}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"type":"message_stop"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

// ==================== ClaudeRequest 测试 ====================

func TestClaudeRequest_Deserialization(t *testing.T) {
	jsonStr := `{
		"max_tokens": 1024,
		"messages": [
			{"role": "user", "content": "Hello"}
		],
		"stream": true,
		"system": "You are helpful.",
		"tools": [{"name": "test", "description": "A test tool"}]
	}`

	var req ClaudeRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.MaxTokens != 1024 {
		t.Errorf("expected max_tokens 1024")
	}
	if len(req.Messages) != 1 {
		t.Errorf("expected 1 message")
	}
	if !req.Stream {
		t.Errorf("expected stream true")
	}
	if len(req.System) == 0 {
		t.Errorf("expected system prompt")
	}
	if len(req.Tools) == 0 {
		t.Errorf("expected tools")
	}
}

func TestClaudeRequest_WithArrayContent(t *testing.T) {
	jsonStr := `{
		"max_tokens": 1024,
		"messages": [
			{
				"role": "user",
				"content": [{"type": "text", "text": "Hello"}]
			}
		],
		"stream": true
	}`

	var req ClaudeRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(req.Messages) != 1 {
		t.Errorf("expected 1 message")
	}

	// Content 应该是 RawMessage，可以是字符串或数组
	var blocks []ContentBlock
	if err := json.Unmarshal(req.Messages[0].Content, &blocks); err != nil {
		t.Fatalf("failed to unmarshal content as blocks: %v", err)
	}

	if len(blocks) != 1 {
		t.Errorf("expected 1 content block")
	}
	if blocks[0].Type != "text" {
		t.Errorf("expected text type")
	}
}

// ==================== ParsedToolCall 测试 ====================

func TestParsedToolCall_Serialization(t *testing.T) {
	call := ParsedToolCall{
		Name:  "search",
		ID:    "toolu_123",
		Input: json.RawMessage(`{"query": "test"}`),
	}

	data, err := json.Marshal(call)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["name"] != "search" {
		t.Errorf("expected name 'search'")
	}
	if result["id"] != "toolu_123" {
		t.Errorf("expected id 'toolu_123'")
	}
}

func TestParsedToolCall_Deserialization(t *testing.T) {
	jsonStr := `{"name": "calculator", "id": "toolu_abc", "input": {"expression": "1+1"}}`

	var call ParsedToolCall
	if err := json.Unmarshal([]byte(jsonStr), &call); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if call.Name != "calculator" {
		t.Errorf("expected name 'calculator'")
	}
	if call.ID != "toolu_abc" {
		t.Errorf("expected id 'toolu_abc'")
	}

	var input map[string]string
	if err := json.Unmarshal(call.Input, &input); err != nil {
		t.Fatalf("failed to unmarshal input: %v", err)
	}
	if input["expression"] != "1+1" {
		t.Errorf("expected expression '1+1'")
	}
}
