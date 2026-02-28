package claude

import (
	"encoding/json"
	"fmt"

	"puter2api/internal/types"
)

// GetMessageText 获取消息文本内容
func GetMessageText(m *types.ClaudeMessage) string {
	var text string
	if err := json.Unmarshal(m.Content, &text); err == nil {
		return text
	}

	var blocks []types.ContentBlock
	if err := json.Unmarshal(m.Content, &blocks); err == nil {
		var result string
		for _, b := range blocks {
			switch b.Type {
			case "text":
				result += b.Text
			case "tool_use":
				inputStr, _ := json.Marshal(b.Input)
				result += fmt.Sprintf("\n<tool_call>\n{\"name\": \"%s\", \"id\": \"%s\", \"input\": %s}\n</tool_call>\n", b.Name, b.ID, string(inputStr))
			case "tool_result":
				var contentStr string
				if err := json.Unmarshal(b.Content, &contentStr); err != nil {
					contentStr = string(b.Content)
				}
				result += fmt.Sprintf("\n<tool_result id=\"%s\">\n%s\n</tool_result>\n", b.ToolUseID, contentStr)
			}
		}
		return result
	}
	return ""
}

// BuildSystemPrompt 构建包含工具定义的 system prompt
func BuildSystemPrompt(originalSystem json.RawMessage, tools json.RawMessage) string {
	systemText := "You are currently using model: claude-opus-4-5-20251101\n\n"

	if len(originalSystem) > 0 {
		var sysStr string
		if err := json.Unmarshal(originalSystem, &sysStr); err == nil {
			systemText += sysStr
		} else {
			var sysBlocks []types.ContentBlock
			if err := json.Unmarshal(originalSystem, &sysBlocks); err == nil {
				for _, b := range sysBlocks {
					if b.Type == "text" {
						systemText += b.Text + "\n"
					}
				}
			}
		}
	}

	if len(tools) > 0 {
		var toolDefs []types.ToolDef
		if err := json.Unmarshal(tools, &toolDefs); err == nil && len(toolDefs) > 0 {
			toolPrompt := "\n\n# Tools\n\nYou have access to the following tools. When you need to use a tool, output it in this EXACT format:\n\n<tool_call>\n{\"name\": \"tool_name\", \"input\": {\"param\": \"value\"}}\n</tool_call>\n\nAvailable tools:\n\n"
			for _, tool := range toolDefs {
				toolPrompt += fmt.Sprintf("## %s\n", tool.Name)
				if tool.Description != "" {
					toolPrompt += fmt.Sprintf("%s\n", tool.Description)
				}
				if len(tool.InputSchema) > 0 {
					toolPrompt += fmt.Sprintf("Input schema: %s\n", string(tool.InputSchema))
				}
				toolPrompt += "\n"
			}
			systemText += toolPrompt
		}
	}

	return systemText
}

// 上下文字符限制（约等于 100k tokens，按 4 字符/token 估算）
const MaxContextChars = 700000

// ConvertMessages 转换 Claude 消息为 Puter 消息，并在超出限制时截断旧消息
func ConvertMessages(messages []types.ClaudeMessage, systemPrompt string) []types.PuterMessage {
	var result []types.PuterMessage

	// 先添加 system prompt
	if systemPrompt != "" {
		result = append(result, types.PuterMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// 转换所有消息
	var allMessages []types.PuterMessage
	for _, m := range messages {
		allMessages = append(allMessages, types.PuterMessage{
			Role:    m.Role,
			Content: GetMessageText(&m),
		})
	}

	// 计算 system prompt 占用的字符数
	usedChars := len(systemPrompt)

	// 从后往前累加，保留最新的消息
	var keptMessages []types.PuterMessage
	for i := len(allMessages) - 1; i >= 0; i-- {
		msgLen := len(allMessages[i].Content)
		if usedChars+msgLen > MaxContextChars {
			// 超出限制，停止添加更早的消息
			break
		}
		usedChars += msgLen
		keptMessages = append([]types.PuterMessage{allMessages[i]}, keptMessages...)
	}

	// 确保第一条消息是 user 角色（Claude API 要求）
	for len(keptMessages) > 0 && keptMessages[0].Role != "user" {
		keptMessages = keptMessages[1:]
	}

	result = append(result, keptMessages...)
	return result
}
