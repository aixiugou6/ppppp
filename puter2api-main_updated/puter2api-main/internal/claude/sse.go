package claude

import (
	"encoding/json"
	"fmt"

	"puter2api/internal/types"

	"github.com/gin-gonic/gin"
)

// SSEWriter SSE 写入器
type SSEWriter struct {
	c *gin.Context
}

// NewSSEWriter 创建 SSE 写入器
func NewSSEWriter(c *gin.Context) *SSEWriter {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	return &SSEWriter{c: c}
}

// SendEvent 发送 SSE 事件
func (w *SSEWriter) SendEvent(event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	w.c.Writer.WriteString(fmt.Sprintf("event: %s\ndata: %s\n\n", event, jsonData))
	w.c.Writer.Flush()
}

// SendMessageStart 发送 message_start 事件
func (w *SSEWriter) SendMessageStart(msgID, model string) {
	w.SendEvent("message_start", types.MessageStartEvent{
		Type: "message_start",
		Message: types.MessageStartDetail{
			ID:      msgID,
			Type:    "message",
			Role:    "assistant",
			Content: []types.ContentBlock{},
			Model:   model,
			Usage:   types.Usage{InputTokens: 100, OutputTokens: 0},
		},
	})
}

// SendTextBlockStart 发送文本块开始事件
func (w *SSEWriter) SendTextBlockStart(index int) {
	w.SendEvent("content_block_start", types.ContentBlockStartEvent{
		Type:         "content_block_start",
		Index:        index,
		ContentBlock: types.TextContentBlock{Type: "text", Text: ""},
	})
}

// SendTextDelta 发送文本增量
func (w *SSEWriter) SendTextDelta(index int, text string) {
	w.SendEvent("content_block_delta", types.ContentBlockDeltaEvent{
		Type:  "content_block_delta",
		Index: index,
		Delta: types.TextDelta{Type: "text_delta", Text: text},
	})
}

// SendToolUseBlockStart 发送工具使用块开始事件
func (w *SSEWriter) SendToolUseBlockStart(index int, id, name string) {
	w.SendEvent("content_block_start", types.ContentBlockStartEvent{
		Type:  "content_block_start",
		Index: index,
		ContentBlock: types.ToolUseContentBlock{
			Type:  "tool_use",
			ID:    id,
			Name:  name,
			Input: json.RawMessage("{}"),
		},
	})
}

// SendInputJSONDelta 发送 JSON 输入增量
func (w *SSEWriter) SendInputJSONDelta(index int, partialJSON string) {
	w.SendEvent("content_block_delta", types.ContentBlockDeltaEvent{
		Type:  "content_block_delta",
		Index: index,
		Delta: types.InputJSONDelta{Type: "input_json_delta", PartialJSON: partialJSON},
	})
}

// SendBlockStop 发送块结束事件
func (w *SSEWriter) SendBlockStop(index int) {
	w.SendEvent("content_block_stop", types.ContentBlockStopEvent{
		Type:  "content_block_stop",
		Index: index,
	})
}

// SendMessageDelta 发送消息增量事件
func (w *SSEWriter) SendMessageDelta(stopReason string, outputTokens int) {
	w.SendEvent("message_delta", types.MessageDeltaEvent{
		Type:  "message_delta",
		Delta: types.MessageDelta{StopReason: stopReason},
		Usage: types.DeltaUsage{OutputTokens: outputTokens},
	})
}

// SendMessageStop 发送消息结束事件
func (w *SSEWriter) SendMessageStop() {
	w.SendEvent("message_stop", types.MessageStopEvent{Type: "message_stop"})
}
