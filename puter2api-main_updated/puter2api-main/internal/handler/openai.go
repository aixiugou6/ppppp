package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"puter2api/internal/claude"
	"puter2api/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// HandleOpenAIChat 处理 /v1/chat/completions 请求 (OpenAI 兼容接口)
func (h *Handler) HandleOpenAIChat(c *gin.Context) {
	startTime := time.Now()

	// 读取原始请求体用于调试
	bodyBytes, _ := c.GetRawData()

	// 重新设置请求体以便后续解析
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req types.OpenAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Str("api", "OpenAI").Err(err).Msg("JSON 解析失败")
		c.JSON(400, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	hasTools := len(req.Tools) > 0
	lastMsgLen := len(req.Messages[len(req.Messages)-1].Content)
	log.Info().
		Str("api", "OpenAI").
		Str("model", req.Model).
		Bool("stream", req.Stream).
		Int("messages", len(req.Messages)).
		Bool("hasTools", hasTools).
		Int("last_msg_len", lastMsgLen).
		Msg("收到请求")

	// 从数据库获取可用的 Token
	tokenRecord, err := h.store.GetActiveToken()
	if err != nil {
		log.Error().Str("api", "OpenAI").Err(err).Msg("获取 Token 失败")
		c.JSON(500, gin.H{
			"error": gin.H{
				"message": "failed to get token",
				"type":    "api_error",
				"code":    "internal_error",
			},
		})
		return
	}
	if tokenRecord == nil {
		c.JSON(401, gin.H{
			"error": gin.H{
				"message": "no active token available, please add a token first",
				"type":    "authentication_error",
				"code":    "invalid_api_key",
			},
		})
		return
	}

	token := tokenRecord.Token
	log.Debug().Str("api", "OpenAI").Str("token", tokenRecord.Name).Int64("id", tokenRecord.ID).Msg("使用 Token")

	// 更新 Token 使用时间
	h.store.UpdateTokenUsed(tokenRecord.ID)

	// 转换 OpenAI 消息为 Puter 消息
	systemPrompt, messages := h.convertOpenAIMessages(req)
	puterMessages := claude.ConvertMessages(messages, systemPrompt)

	// 调用 Puter API
	responseText, err := h.puterClient.Call(puterMessages, token)
	if err != nil {
		log.Error().Str("api", "OpenAI").Err(err).Msg("调用 Puter API 失败")
		c.JSON(500, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "api_error",
				"code":    "internal_error",
			},
		})
		return
	}

	// 解析工具调用
	toolCalls, remainingText := claude.ParseToolCalls(responseText)

	// 发送响应
	if req.Stream {
		h.sendOpenAIStreamResponse(c, req.Model, remainingText, toolCalls)
	} else {
		h.sendOpenAINonStreamResponse(c, req.Model, remainingText, toolCalls)
	}

	// 记录完成日志
	elapsed := time.Since(startTime).Seconds()
	log.Info().
		Str("api", "OpenAI").
		Str("耗时", fmt.Sprintf("%.2fs", elapsed)).
		Int("响应长度", len(responseText)).
		Msg("请求完成")
}

// convertOpenAIMessages 转换 OpenAI 消息格式为内部格式
func (h *Handler) convertOpenAIMessages(req types.OpenAIRequest) (string, []types.ClaudeMessage) {
	var systemPrompt string
	var messages []types.ClaudeMessage

	// 处理工具定义，添加到 system prompt
	if len(req.Tools) > 0 {
		toolPrompt := "\n\n# Tools\n\nYou have access to the following tools. When you need to use a tool, output it in this EXACT format:\n\n<tool_call>\n{\"name\": \"tool_name\", \"input\": {\"param\": \"value\"}}\n</tool_call>\n\nAvailable tools:\n\n"
		for _, tool := range req.Tools {
			toolPrompt += fmt.Sprintf("## %s\n", tool.Function.Name)
			if tool.Function.Description != "" {
				toolPrompt += fmt.Sprintf("%s\n", tool.Function.Description)
			}
			if len(tool.Function.Parameters) > 0 {
				toolPrompt += fmt.Sprintf("Input schema: %s\n", string(tool.Function.Parameters))
			}
			toolPrompt += "\n"
		}
		systemPrompt = toolPrompt
	}

	for _, m := range req.Messages {
		if m.Role == "system" {
			// 提取 system 消息内容
			var content string
			if err := json.Unmarshal(m.Content, &content); err == nil {
				systemPrompt = content + systemPrompt
			} else {
				systemPrompt = string(m.Content) + systemPrompt
			}
			continue
		}

		// 转换消息
		claudeMsg := types.ClaudeMessage{
			Role: m.Role,
		}

		// 处理 tool 角色的消息
		if m.Role == "tool" {
			claudeMsg.Role = "user"
			var content string
			if err := json.Unmarshal(m.Content, &content); err != nil {
				content = string(m.Content)
			}
			toolResult := fmt.Sprintf("\n<tool_result id=\"%s\">\n%s\n</tool_result>\n", m.ToolCallID, content)
			claudeMsg.Content, _ = json.Marshal(toolResult)
		} else if len(m.ToolCalls) > 0 {
			// 处理 assistant 消息中的 tool_calls
			var content string
			if err := json.Unmarshal(m.Content, &content); err == nil && content != "" {
				// 有文本内容
			} else {
				content = ""
			}
			for _, tc := range m.ToolCalls {
				content += fmt.Sprintf("\n<tool_call>\n{\"name\": \"%s\", \"id\": \"%s\", \"input\": %s}\n</tool_call>\n", tc.Function.Name, tc.ID, tc.Function.Arguments)
			}
			claudeMsg.Content, _ = json.Marshal(content)
		} else {
			// 普通消息
			claudeMsg.Content = m.Content
		}

		messages = append(messages, claudeMsg)
	}

	return systemPrompt, messages
}

// sendOpenAIStreamResponse 发送 OpenAI 格式的流式响应
func (h *Handler) sendOpenAIStreamResponse(c *gin.Context, model string, text string, toolCalls []types.ParsedToolCall) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	msgID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()

	// 发送角色信息
	firstChunk := types.OpenAIResponse{
		ID:      msgID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []types.OpenAIChoice{
			{
				Index: 0,
				Delta: &types.OpenAIResponseMsg{
					Role: "assistant",
				},
				FinishReason: nil,
				Logprobs:     nil,
			},
		},
	}
	h.writeSSEChunk(c, firstChunk)

	// 发送文本内容（一次性发送）
	if text != "" {
		chunk := types.OpenAIResponse{
			ID:      msgID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   model,
			Choices: []types.OpenAIChoice{
				{
					Index: 0,
					Delta: &types.OpenAIResponseMsg{
						Content: &text,
					},
					FinishReason: nil,
					Logprobs:     nil,
				},
			},
		}
		h.writeSSEChunk(c, chunk)
	}

	// 发送工具调用
	if len(toolCalls) > 0 {
		for i, tc := range toolCalls {
			idx := i
			// 发送工具调用开始
			toolCallChunk := types.OpenAIResponse{
				ID:      msgID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   model,
				Choices: []types.OpenAIChoice{
					{
						Index: 0,
						Delta: &types.OpenAIResponseMsg{
							ToolCalls: []types.OpenAIToolCall{
								{
									Index: &idx,
									ID:    tc.ID,
									Type:  "function",
									Function: types.OpenAIToolCallFunction{
										Name:      tc.Name,
										Arguments: "",
									},
								},
							},
						},
						FinishReason: nil,
						Logprobs:     nil,
					},
				},
			}
			h.writeSSEChunk(c, toolCallChunk)

			// 一次性发送参数
			argsStr := string(tc.Input)
			argChunk := types.OpenAIResponse{
				ID:      msgID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   model,
				Choices: []types.OpenAIChoice{
					{
						Index: 0,
						Delta: &types.OpenAIResponseMsg{
							ToolCalls: []types.OpenAIToolCall{
								{
									Index: &idx,
									Function: types.OpenAIToolCallFunction{
										Arguments: argsStr,
									},
								},
							},
						},
						FinishReason: nil,
						Logprobs:     nil,
					},
				},
			}
			h.writeSSEChunk(c, argChunk)
		}
	}

	// 发送结束标记
	finishReason := "stop"
	if len(toolCalls) > 0 {
		finishReason = "tool_calls"
	}
	finalChunk := types.OpenAIResponse{
		ID:      msgID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []types.OpenAIChoice{
			{
				Index:        0,
				Delta:        &types.OpenAIResponseMsg{},
				FinishReason: &finishReason,
				Logprobs:     nil,
			},
		},
	}
	h.writeSSEChunk(c, finalChunk)

	// 发送 [DONE]
	c.Writer.Write([]byte("data: [DONE]\n\n"))
	c.Writer.Flush()
}

// sendOpenAINonStreamResponse 发送 OpenAI 格式的非流式响应
func (h *Handler) sendOpenAINonStreamResponse(c *gin.Context, model string, text string, toolCalls []types.ParsedToolCall) {
	msgID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()

	finishReason := "stop"
	if len(toolCalls) > 0 {
		finishReason = "tool_calls"
	}

	var openaiToolCalls []types.OpenAIToolCall
	for _, tc := range toolCalls {
		openaiToolCalls = append(openaiToolCalls, types.OpenAIToolCall{
			ID:   tc.ID,
			Type: "function",
			Function: types.OpenAIToolCallFunction{
				Name:      tc.Name,
				Arguments: string(tc.Input),
			},
		})
	}

	var contentPtr *string
	if text != "" {
		contentPtr = &text
	}

	resp := types.OpenAIResponse{
		ID:      msgID,
		Object:  "chat.completion",
		Created: created,
		Model:   model,
		Choices: []types.OpenAIChoice{
			{
				Index: 0,
				Message: &types.OpenAIResponseMsg{
					Role:      "assistant",
					Content:   contentPtr,
					ToolCalls: openaiToolCalls,
				},
				FinishReason: &finishReason,
				Logprobs:     nil,
			},
		},
		Usage: &types.OpenAIUsage{
			PromptTokens:     0,
			CompletionTokens: len(text),
			TotalTokens:      len(text),
		},
	}

	c.JSON(200, resp)
}

// writeSSEChunk 写入 SSE 数据块
func (h *Handler) writeSSEChunk(c *gin.Context, data any) {
	jsonData, _ := json.Marshal(data)
	c.Writer.Write([]byte("data: "))
	c.Writer.Write(jsonData)
	c.Writer.Write([]byte("\n\n"))
	c.Writer.Flush()
}

// HandleModels 处理 /v1/models 请求
func (h *Handler) HandleModels(c *gin.Context) {
	// Prefer model.json (ships with the repo) so the /v1/models response stays up-to-date.
	// If the file isn't present (or can't be parsed), fall back to the original hardcoded list.
	ids, err := loadModelIDs()
	var models []map[string]any
	if err == nil && len(ids) > 0 {
		created := time.Now().Unix()
		models = make([]map[string]any, 0, len(ids))
		for _, id := range ids {
			models = append(models, map[string]any{
				"id":       id,
				"object":   "model",
				"created":  created,
				"owned_by": inferOwner(id),
			})
		}
	} else {
		models = []map[string]any{
			{
				"id":       "claude-opus-4-5",
				"object":   "model",
				"created":  1700000000,
				"owned_by": "anthropic",
			},
			{
				"id":       "gpt-4",
				"object":   "model",
				"created":  1700000000,
				"owned_by": "openai",
			},
			{
				"id":       "gpt-4-turbo",
				"object":   "model",
				"created":  1700000000,
				"owned_by": "openai",
			},
			{
				"id":       "gpt-3.5-turbo",
				"object":   "model",
				"created":  1700000000,
				"owned_by": "openai",
			},
		}
	}

	c.JSON(200, gin.H{
		"object": "list",
		"data":   models,
	})
}
