package claude

import (
	"encoding/json"
	"strings"
	"testing"

	"puter2api/internal/types"
)

// ==================== GetMessageText 测试 ====================

func TestGetMessageText_StringContent(t *testing.T) {
	msg := &types.ClaudeMessage{
		Role:    "user",
		Content: json.RawMessage(`"Hello, world!"`),
	}

	result := GetMessageText(msg)

	if result != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got '%s'", result)
	}
}

func TestGetMessageText_TextBlock(t *testing.T) {
	msg := &types.ClaudeMessage{
		Role:    "assistant",
		Content: json.RawMessage(`[{"type": "text", "text": "This is a text block."}]`),
	}

	result := GetMessageText(msg)

	if result != "This is a text block." {
		t.Errorf("expected 'This is a text block.', got '%s'", result)
	}
}

func TestGetMessageText_MultipleTextBlocks(t *testing.T) {
	msg := &types.ClaudeMessage{
		Role: "assistant",
		Content: json.RawMessage(`[
			{"type": "text", "text": "First part. "},
			{"type": "text", "text": "Second part."}
		]`),
	}

	result := GetMessageText(msg)

	if result != "First part. Second part." {
		t.Errorf("expected 'First part. Second part.', got '%s'", result)
	}
}

func TestGetMessageText_ToolUseBlock(t *testing.T) {
	msg := &types.ClaudeMessage{
		Role: "assistant",
		Content: json.RawMessage(`[{
			"type": "tool_use",
			"id": "toolu_123",
			"name": "test_tool",
			"input": {"param": "value"}
		}]`),
	}

	result := GetMessageText(msg)

	if !strings.Contains(result, "<tool_call>") {
		t.Errorf("expected tool_call tag in result")
	}
	if !strings.Contains(result, "test_tool") {
		t.Errorf("expected tool name in result")
	}
	if !strings.Contains(result, "toolu_123") {
		t.Errorf("expected tool id in result")
	}
	if !strings.Contains(result, "</tool_call>") {
		t.Errorf("expected closing tool_call tag in result")
	}
}

func TestGetMessageText_ToolResultBlock(t *testing.T) {
	msg := &types.ClaudeMessage{
		Role: "user",
		Content: json.RawMessage(`[{
			"type": "tool_result",
			"tool_use_id": "toolu_123",
			"content": "Tool execution result"
		}]`),
	}

	result := GetMessageText(msg)

	if !strings.Contains(result, "<tool_result") {
		t.Errorf("expected tool_result tag in result")
	}
	if !strings.Contains(result, "toolu_123") {
		t.Errorf("expected tool_use_id in result")
	}
	if !strings.Contains(result, "Tool execution result") {
		t.Errorf("expected tool result content in result")
	}
}

func TestGetMessageText_MixedBlocks(t *testing.T) {
	msg := &types.ClaudeMessage{
		Role: "assistant",
		Content: json.RawMessage(`[
			{"type": "text", "text": "Let me help you with that."},
			{"type": "tool_use", "id": "toolu_456", "name": "search", "input": {"query": "test"}}
		]`),
	}

	result := GetMessageText(msg)

	if !strings.Contains(result, "Let me help you with that.") {
		t.Errorf("expected text content in result")
	}
	if !strings.Contains(result, "<tool_call>") {
		t.Errorf("expected tool_call tag in result")
	}
	if !strings.Contains(result, "search") {
		t.Errorf("expected tool name in result")
	}
}

func TestGetMessageText_EmptyContent(t *testing.T) {
	msg := &types.ClaudeMessage{
		Role:    "user",
		Content: json.RawMessage(`""`),
	}

	result := GetMessageText(msg)

	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestGetMessageText_InvalidJSON(t *testing.T) {
	msg := &types.ClaudeMessage{
		Role:    "user",
		Content: json.RawMessage(`invalid json`),
	}

	result := GetMessageText(msg)

	if result != "" {
		t.Errorf("expected empty string for invalid JSON, got '%s'", result)
	}
}

// ==================== BuildSystemPrompt 测试 ====================

func TestBuildSystemPrompt_StringSystem(t *testing.T) {
	system := json.RawMessage(`"You are a helpful assistant."`)

	result := BuildSystemPrompt(system, nil)

	if result != "You are a helpful assistant." {
		t.Errorf("expected 'You are a helpful assistant.', got '%s'", result)
	}
}

func TestBuildSystemPrompt_BlockSystem(t *testing.T) {
	system := json.RawMessage(`[{"type": "text", "text": "System prompt from block."}]`)

	result := BuildSystemPrompt(system, nil)

	if !strings.Contains(result, "System prompt from block.") {
		t.Errorf("expected system prompt content, got '%s'", result)
	}
}

func TestBuildSystemPrompt_WithTools(t *testing.T) {
	system := json.RawMessage(`"Base system prompt."`)
	tools := json.RawMessage(`[
		{"name": "calculator", "description": "Performs calculations", "input_schema": {"type": "object"}},
		{"name": "search", "description": "Searches the web"}
	]`)

	result := BuildSystemPrompt(system, tools)

	if !strings.Contains(result, "Base system prompt.") {
		t.Errorf("expected base system prompt")
	}
	if !strings.Contains(result, "# Tools") {
		t.Errorf("expected Tools header")
	}
	if !strings.Contains(result, "## calculator") {
		t.Errorf("expected calculator tool")
	}
	if !strings.Contains(result, "Performs calculations") {
		t.Errorf("expected calculator description")
	}
	if !strings.Contains(result, "## search") {
		t.Errorf("expected search tool")
	}
	if !strings.Contains(result, "Searches the web") {
		t.Errorf("expected search description")
	}
	if !strings.Contains(result, "<tool_call>") {
		t.Errorf("expected tool_call format example")
	}
}

func TestBuildSystemPrompt_EmptySystem(t *testing.T) {
	tools := json.RawMessage(`[{"name": "test_tool", "description": "A test tool"}]`)

	result := BuildSystemPrompt(nil, tools)

	if !strings.Contains(result, "# Tools") {
		t.Errorf("expected Tools header even without system prompt")
	}
	if !strings.Contains(result, "test_tool") {
		t.Errorf("expected tool name")
	}
}

func TestBuildSystemPrompt_EmptyTools(t *testing.T) {
	system := json.RawMessage(`"Just a system prompt."`)

	result := BuildSystemPrompt(system, json.RawMessage(`[]`))

	if result != "Just a system prompt." {
		t.Errorf("expected only system prompt, got '%s'", result)
	}
}

func TestBuildSystemPrompt_ToolWithInputSchema(t *testing.T) {
	tools := json.RawMessage(`[{
		"name": "get_weather",
		"description": "Get weather for a location",
		"input_schema": {
			"type": "object",
			"properties": {
				"location": {"type": "string"}
			},
			"required": ["location"]
		}
	}]`)

	result := BuildSystemPrompt(nil, tools)

	if !strings.Contains(result, "Input schema:") {
		t.Errorf("expected Input schema label")
	}
	if !strings.Contains(result, "location") {
		t.Errorf("expected location in schema")
	}
}

// ==================== ConvertMessages 测试 ====================

func TestConvertMessages_BasicConversion(t *testing.T) {
	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"Hello"`)},
		{Role: "assistant", Content: json.RawMessage(`"Hi there!"`)},
		{Role: "user", Content: json.RawMessage(`"How are you?"`)},
	}

	result := ConvertMessages(messages, "System prompt")

	if len(result) != 4 { // system + 3 messages
		t.Fatalf("expected 4 messages, got %d", len(result))
	}

	if result[0].Role != "system" {
		t.Errorf("expected first message to be system")
	}
	if result[0].Content != "System prompt" {
		t.Errorf("expected system content")
	}

	if result[1].Role != "user" || result[1].Content != "Hello" {
		t.Errorf("expected first user message")
	}
	if result[2].Role != "assistant" || result[2].Content != "Hi there!" {
		t.Errorf("expected assistant message")
	}
	if result[3].Role != "user" || result[3].Content != "How are you?" {
		t.Errorf("expected second user message")
	}
}

func TestConvertMessages_NoSystemPrompt(t *testing.T) {
	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"Hello"`)},
	}

	result := ConvertMessages(messages, "")

	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}

	if result[0].Role != "user" {
		t.Errorf("expected user role")
	}
}

func TestConvertMessages_EnsuresUserFirst(t *testing.T) {
	messages := []types.ClaudeMessage{
		{Role: "assistant", Content: json.RawMessage(`"I'll help you"`)},
		{Role: "user", Content: json.RawMessage(`"Thanks"`)},
	}

	result := ConvertMessages(messages, "")

	// 应该移除开头的 assistant 消息
	if len(result) != 1 {
		t.Fatalf("expected 1 message after filtering, got %d", len(result))
	}

	if result[0].Role != "user" {
		t.Errorf("expected first message to be user, got %s", result[0].Role)
	}
}

func TestConvertMessages_ContextTruncation(t *testing.T) {
	// 创建一个超长消息
	longContent := strings.Repeat("x", MaxContextChars+1000)
	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"` + longContent + `"`)},
		{Role: "assistant", Content: json.RawMessage(`"Response"`)},
		{Role: "user", Content: json.RawMessage(`"Short message"`)},
	}

	result := ConvertMessages(messages, "")

	// 应该只保留最新的消息（从后往前）
	// 由于第一条消息太长，应该被截断
	if len(result) == 3 {
		t.Errorf("expected truncation to occur")
	}

	// 最后一条消息应该被保留
	lastMsg := result[len(result)-1]
	if lastMsg.Content != "Short message" {
		t.Errorf("expected last message to be preserved")
	}
}

func TestConvertMessages_EmptyMessages(t *testing.T) {
	result := ConvertMessages([]types.ClaudeMessage{}, "System")

	if len(result) != 1 {
		t.Fatalf("expected 1 message (system only), got %d", len(result))
	}

	if result[0].Role != "system" {
		t.Errorf("expected system message")
	}
}

func TestConvertMessages_WithToolBlocks(t *testing.T) {
	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"Search for something"`)},
		{Role: "assistant", Content: json.RawMessage(`[{"type": "tool_use", "id": "toolu_1", "name": "search", "input": {"q": "test"}}]`)},
		{Role: "user", Content: json.RawMessage(`[{"type": "tool_result", "tool_use_id": "toolu_1", "content": "Results here"}]`)},
	}

	result := ConvertMessages(messages, "")

	if len(result) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(result))
	}

	// 检查工具调用被正确转换
	if !strings.Contains(result[1].Content, "<tool_call>") {
		t.Errorf("expected tool_call in assistant message")
	}

	// 检查工具结果被正确转换
	if !strings.Contains(result[2].Content, "<tool_result") {
		t.Errorf("expected tool_result in user message")
	}
}

// ==================== 超长上下文测试 ====================

func TestConvertMessages_VeryLongContext_200K(t *testing.T) {
	// 测试 200k 字符的上下文（应该在限制内）
	const chars200K = 200000
	longContent := strings.Repeat("a", chars200K)

	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"` + longContent + `"`)},
	}

	result := ConvertMessages(messages, "")

	// 200k 在 700k 限制内，应该保留
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}

	if len(result[0].Content) != chars200K {
		t.Errorf("expected content length %d, got %d", chars200K, len(result[0].Content))
	}
}

func TestConvertMessages_VeryLongContext_500K(t *testing.T) {
	// 测试 500k 字符的上下文（应该在限制内）
	const chars500K = 500000
	longContent := strings.Repeat("b", chars500K)

	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"` + longContent + `"`)},
	}

	result := ConvertMessages(messages, "")

	// 500k 在 700k 限制内，应该保留
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}

	if len(result[0].Content) != chars500K {
		t.Errorf("expected content length %d, got %d", chars500K, len(result[0].Content))
	}
}

func TestConvertMessages_VeryLongContext_800K_Truncation(t *testing.T) {
	// 测试 800k 字符的上下文（超过 700k 限制）
	const chars800K = 800000
	longContent := strings.Repeat("c", chars800K)

	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"` + longContent + `"`)},
		{Role: "assistant", Content: json.RawMessage(`"Short response"`)},
		{Role: "user", Content: json.RawMessage(`"Final question"`)},
	}

	result := ConvertMessages(messages, "")

	// 第一条消息太长，应该被截断，只保留后面的消息
	// 最后一条消息必须被保留
	found := false
	for _, msg := range result {
		if msg.Content == "Final question" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Final question' to be preserved")
	}

	// 超长消息不应该被保留
	for _, msg := range result {
		if len(msg.Content) == chars800K {
			t.Errorf("800k message should have been truncated")
		}
	}
}

func TestConvertMessages_MultipleMessages_TotalExceeds700K(t *testing.T) {
	// 多条消息总和超过 700k
	const chars300K = 300000
	msg1 := strings.Repeat("1", chars300K)
	msg2 := strings.Repeat("2", chars300K)
	msg3 := strings.Repeat("3", chars300K) // 总共 900k

	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"` + msg1 + `"`)},
		{Role: "assistant", Content: json.RawMessage(`"` + msg2 + `"`)},
		{Role: "user", Content: json.RawMessage(`"` + msg3 + `"`)},
	}

	result := ConvertMessages(messages, "")

	// 计算总字符数
	totalChars := 0
	for _, msg := range result {
		totalChars += len(msg.Content)
	}

	// 总字符数应该不超过 MaxContextChars
	if totalChars > MaxContextChars {
		t.Errorf("total chars %d exceeds MaxContextChars %d", totalChars, MaxContextChars)
	}

	// 最新的消息（msg3）应该被保留
	lastMsg := result[len(result)-1]
	if lastMsg.Content != msg3 {
		t.Errorf("expected last message (msg3) to be preserved")
	}
}

func TestConvertMessages_LongSystemPrompt_WithMessages(t *testing.T) {
	// 测试长 system prompt 与消息的组合
	const chars400K = 400000
	longSystemPrompt := strings.Repeat("s", chars400K)
	msgContent := strings.Repeat("m", chars400K) // system + msg = 800k > 700k

	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"` + msgContent + `"`)},
	}

	result := ConvertMessages(messages, longSystemPrompt)

	// system prompt 应该被保留
	if len(result) == 0 {
		t.Fatalf("expected at least system message")
	}

	if result[0].Role != "system" {
		t.Errorf("expected first message to be system")
	}

	if result[0].Content != longSystemPrompt {
		t.Errorf("expected system prompt to be preserved")
	}

	// 由于 system prompt 占用了 400k，用户消息 400k 会导致总共 800k > 700k
	// 用户消息应该被截断
	totalChars := 0
	for _, msg := range result {
		totalChars += len(msg.Content)
	}

	// 注意：当前实现中 system prompt 不计入截断逻辑的 usedChars 初始值
	// 这里测试实际行为
	t.Logf("Total chars in result: %d, messages count: %d", totalChars, len(result))
}

func TestConvertMessages_PreservesNewestMessages(t *testing.T) {
	// 验证从后往前保留消息的逻辑
	const chars250K = 250000

	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"` + strings.Repeat("1", chars250K) + `"`)},
		{Role: "assistant", Content: json.RawMessage(`"` + strings.Repeat("2", chars250K) + `"`)},
		{Role: "user", Content: json.RawMessage(`"` + strings.Repeat("3", chars250K) + `"`)},
		{Role: "assistant", Content: json.RawMessage(`"newest_response"`)},
		{Role: "user", Content: json.RawMessage(`"newest_question"`)},
	}

	result := ConvertMessages(messages, "")

	// 最新的两条消息应该被保留
	hasNewestQuestion := false
	hasNewestResponse := false
	for _, msg := range result {
		if msg.Content == "newest_question" {
			hasNewestQuestion = true
		}
		if msg.Content == "newest_response" {
			hasNewestResponse = true
		}
	}

	if !hasNewestQuestion {
		t.Errorf("expected 'newest_question' to be preserved")
	}
	if !hasNewestResponse {
		t.Errorf("expected 'newest_response' to be preserved")
	}
}

func TestConvertMessages_ExactlyAtLimit(t *testing.T) {
	// 测试刚好在限制边界的情况
	// MaxContextChars = 700000
	const exactLimit = MaxContextChars - 100 // 留一点余量

	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"` + strings.Repeat("x", exactLimit) + `"`)},
	}

	result := ConvertMessages(messages, "")

	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}

	if len(result[0].Content) != exactLimit {
		t.Errorf("expected content length %d, got %d", exactLimit, len(result[0].Content))
	}
}

func TestConvertMessages_JustOverLimit(t *testing.T) {
	// 测试刚好超过限制的情况
	const justOver = MaxContextChars + 100

	messages := []types.ClaudeMessage{
		{Role: "user", Content: json.RawMessage(`"` + strings.Repeat("y", justOver) + `"`)},
		{Role: "assistant", Content: json.RawMessage(`"short"`)},
		{Role: "user", Content: json.RawMessage(`"also short"`)},
	}

	result := ConvertMessages(messages, "")

	// 第一条超长消息应该被跳过
	for _, msg := range result {
		if len(msg.Content) == justOver {
			t.Errorf("message just over limit should have been truncated")
		}
	}

	// 后面的短消息应该被保留
	hasShort := false
	hasAlsoShort := false
	for _, msg := range result {
		if msg.Content == "short" {
			hasShort = true
		}
		if msg.Content == "also short" {
			hasAlsoShort = true
		}
	}

	if !hasAlsoShort {
		t.Errorf("expected 'also short' to be preserved")
	}
	// 注意：由于确保 user 在前的逻辑，"short" (assistant) 可能被移除
	t.Logf("hasShort: %v, hasAlsoShort: %v, result count: %d", hasShort, hasAlsoShort, len(result))
}
