package claude

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"puter2api/internal/types"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// 创建测试用的 gin.Context
func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// ==================== SSEWriter 基础测试 ====================

func TestNewSSEWriter_SetsHeaders(t *testing.T) {
	c, w := createTestContext()

	_ = NewSSEWriter(c)

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("expected Content-Type 'text/event-stream', got '%s'", contentType)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("expected Cache-Control 'no-cache', got '%s'", cacheControl)
	}
}

func TestSSEWriter_SendEvent(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	testData := map[string]string{"key": "value"}
	sse.SendEvent("test_event", testData)

	body := w.Body.String()

	if !strings.Contains(body, "event: test_event") {
		t.Errorf("expected event name in output")
	}
	if !strings.Contains(body, `"key":"value"`) {
		t.Errorf("expected JSON data in output")
	}
	if !strings.Contains(body, "data: ") {
		t.Errorf("expected data prefix in output")
	}
}

// ==================== MessageStart 测试 ====================

func TestSSEWriter_SendMessageStart(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendMessageStart("msg_123", "claude-3-opus")

	body := w.Body.String()

	if !strings.Contains(body, "event: message_start") {
		t.Errorf("expected message_start event")
	}
	if !strings.Contains(body, `"id":"msg_123"`) {
		t.Errorf("expected message id")
	}
	if !strings.Contains(body, `"model":"claude-3-opus"`) {
		t.Errorf("expected model name")
	}
	if !strings.Contains(body, `"role":"assistant"`) {
		t.Errorf("expected assistant role")
	}
	if !strings.Contains(body, `"type":"message"`) {
		t.Errorf("expected message type")
	}

	// 验证 JSON 结构
	var event types.MessageStartEvent
	jsonStr := extractJSONFromSSE(body)
	if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
		t.Fatalf("failed to parse message_start JSON: %v", err)
	}

	if event.Message.ID != "msg_123" {
		t.Errorf("expected message ID 'msg_123', got '%s'", event.Message.ID)
	}
}

// ==================== TextBlock 测试 ====================

func TestSSEWriter_SendTextBlockStart(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendTextBlockStart(0)

	body := w.Body.String()

	if !strings.Contains(body, "event: content_block_start") {
		t.Errorf("expected content_block_start event")
	}
	if !strings.Contains(body, `"index":0`) {
		t.Errorf("expected index 0")
	}
	if !strings.Contains(body, `"type":"text"`) {
		t.Errorf("expected text type in content_block")
	}

	// 验证 TextContentBlock 结构（text 字段始终存在）
	var event types.ContentBlockStartEvent
	jsonStr := extractJSONFromSSE(body)
	if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
		t.Fatalf("failed to parse content_block_start JSON: %v", err)
	}

	// 验证 content_block 包含 text 字段
	blockJSON, _ := json.Marshal(event.ContentBlock)
	if !strings.Contains(string(blockJSON), `"text"`) {
		t.Errorf("expected text field in content_block, got: %s", string(blockJSON))
	}
}

func TestSSEWriter_SendTextDelta(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendTextDelta(0, "Hello, world!")

	body := w.Body.String()

	if !strings.Contains(body, "event: content_block_delta") {
		t.Errorf("expected content_block_delta event")
	}
	if !strings.Contains(body, `"type":"text_delta"`) {
		t.Errorf("expected text_delta type")
	}
	if !strings.Contains(body, `"text":"Hello, world!"`) {
		t.Errorf("expected text content")
	}
}

func TestSSEWriter_SendTextDelta_SpecialCharacters(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	// 测试特殊字符
	sse.SendTextDelta(0, "Line1\nLine2\tTabbed\"Quoted\"")

	body := w.Body.String()

	// JSON 应该正确转义
	if !strings.Contains(body, `\n`) {
		t.Errorf("expected escaped newline")
	}
	if !strings.Contains(body, `\t`) {
		t.Errorf("expected escaped tab")
	}
}

// ==================== ToolUseBlock 测试 ====================

func TestSSEWriter_SendToolUseBlockStart(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendToolUseBlockStart(1, "toolu_abc123", "search")

	body := w.Body.String()

	if !strings.Contains(body, "event: content_block_start") {
		t.Errorf("expected content_block_start event")
	}
	if !strings.Contains(body, `"index":1`) {
		t.Errorf("expected index 1")
	}
	if !strings.Contains(body, `"type":"tool_use"`) {
		t.Errorf("expected tool_use type")
	}
	if !strings.Contains(body, `"id":"toolu_abc123"`) {
		t.Errorf("expected tool id")
	}
	if !strings.Contains(body, `"name":"search"`) {
		t.Errorf("expected tool name")
	}

	// 验证 ToolUseContentBlock 结构
	var event types.ContentBlockStartEvent
	jsonStr := extractJSONFromSSE(body)
	if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
		t.Fatalf("failed to parse content_block_start JSON: %v", err)
	}

	// 验证 content_block 包含必要字段
	blockJSON, _ := json.Marshal(event.ContentBlock)
	blockStr := string(blockJSON)
	if !strings.Contains(blockStr, `"input"`) {
		t.Errorf("expected input field in tool_use content_block, got: %s", blockStr)
	}
}

func TestSSEWriter_SendInputJSONDelta(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendInputJSONDelta(1, `{"query": "test"}`)

	body := w.Body.String()

	if !strings.Contains(body, "event: content_block_delta") {
		t.Errorf("expected content_block_delta event")
	}
	if !strings.Contains(body, `"type":"input_json_delta"`) {
		t.Errorf("expected input_json_delta type")
	}
	if !strings.Contains(body, `"partial_json"`) {
		t.Errorf("expected partial_json field")
	}
}

// ==================== BlockStop 测试 ====================

func TestSSEWriter_SendBlockStop(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendBlockStop(2)

	body := w.Body.String()

	if !strings.Contains(body, "event: content_block_stop") {
		t.Errorf("expected content_block_stop event")
	}
	if !strings.Contains(body, `"index":2`) {
		t.Errorf("expected index 2")
	}
}

// ==================== MessageDelta 测试 ====================

func TestSSEWriter_SendMessageDelta_EndTurn(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendMessageDelta("end_turn", 150)

	body := w.Body.String()

	if !strings.Contains(body, "event: message_delta") {
		t.Errorf("expected message_delta event")
	}
	if !strings.Contains(body, `"stop_reason":"end_turn"`) {
		t.Errorf("expected stop_reason end_turn")
	}
	if !strings.Contains(body, `"output_tokens":150`) {
		t.Errorf("expected output_tokens 150")
	}
}

func TestSSEWriter_SendMessageDelta_ToolUse(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendMessageDelta("tool_use", 200)

	body := w.Body.String()

	if !strings.Contains(body, `"stop_reason":"tool_use"`) {
		t.Errorf("expected stop_reason tool_use")
	}
}

// ==================== MessageStop 测试 ====================

func TestSSEWriter_SendMessageStop(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendMessageStop()

	body := w.Body.String()

	if !strings.Contains(body, "event: message_stop") {
		t.Errorf("expected message_stop event")
	}
	if !strings.Contains(body, `"type":"message_stop"`) {
		t.Errorf("expected message_stop type")
	}
}

// ==================== 完整流程测试 ====================

func TestSSEWriter_FullTextResponse(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	// 模拟完整的文本响应流程
	sse.SendMessageStart("msg_test", "claude-3-opus")
	sse.SendTextBlockStart(0)
	sse.SendTextDelta(0, "Hello, ")
	sse.SendTextDelta(0, "world!")
	sse.SendBlockStop(0)
	sse.SendMessageDelta("end_turn", 10)
	sse.SendMessageStop()

	body := w.Body.String()

	// 验证事件顺序
	events := []string{
		"message_start",
		"content_block_start",
		"content_block_delta",
		"content_block_stop",
		"message_delta",
		"message_stop",
	}

	lastIndex := -1
	for _, event := range events {
		index := strings.Index(body, "event: "+event)
		if index == -1 {
			t.Errorf("missing event: %s", event)
			continue
		}
		if index <= lastIndex {
			t.Errorf("event %s appears out of order", event)
		}
		lastIndex = index
	}
}

func TestSSEWriter_FullToolUseResponse(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	// 模拟带工具调用的响应流程
	sse.SendMessageStart("msg_tool", "claude-3-opus")

	// 文本块
	sse.SendTextBlockStart(0)
	sse.SendTextDelta(0, "Let me search for that.")
	sse.SendBlockStop(0)

	// 工具调用块
	sse.SendToolUseBlockStart(1, "toolu_123", "search")
	sse.SendInputJSONDelta(1, `{"query": "test"}`)
	sse.SendBlockStop(1)

	sse.SendMessageDelta("tool_use", 50)
	sse.SendMessageStop()

	body := w.Body.String()

	// 验证包含两个 content_block_start
	count := strings.Count(body, "event: content_block_start")
	if count != 2 {
		t.Errorf("expected 2 content_block_start events, got %d", count)
	}

	// 验证 stop_reason 是 tool_use
	if !strings.Contains(body, `"stop_reason":"tool_use"`) {
		t.Errorf("expected stop_reason tool_use")
	}
}

func TestSSEWriter_MultipleToolCalls(t *testing.T) {
	c, w := createTestContext()
	sse := NewSSEWriter(c)

	sse.SendMessageStart("msg_multi", "claude-3-opus")

	// 多个工具调用
	for i := 0; i < 3; i++ {
		sse.SendToolUseBlockStart(i, "toolu_"+string(rune('a'+i)), "tool_"+string(rune('1'+i)))
		sse.SendInputJSONDelta(i, `{}`)
		sse.SendBlockStop(i)
	}

	sse.SendMessageDelta("tool_use", 100)
	sse.SendMessageStop()

	body := w.Body.String()

	// 验证有 3 个工具调用块
	count := strings.Count(body, `"type":"tool_use"`)
	if count != 3 {
		t.Errorf("expected 3 tool_use blocks, got %d", count)
	}
}

// ==================== 辅助函数 ====================

// extractJSONFromSSE 从 SSE 输出中提取 JSON 数据
func extractJSONFromSSE(sse string) string {
	lines := strings.Split(sse, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			return strings.TrimPrefix(line, "data: ")
		}
	}
	return ""
}
