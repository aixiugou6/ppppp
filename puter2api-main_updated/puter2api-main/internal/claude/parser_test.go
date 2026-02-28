package claude

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseToolCalls_SingleToolCall(t *testing.T) {
	text := `Here is some text before the tool call.
<tool_call>
{"name": "test_tool", "id": "toolu_123", "input": {"param": "value"}}
</tool_call>
And some text after.`

	calls, remaining := ParseToolCalls(text)

	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}

	if calls[0].Name != "test_tool" {
		t.Errorf("expected name 'test_tool', got '%s'", calls[0].Name)
	}

	if calls[0].ID != "toolu_123" {
		t.Errorf("expected id 'toolu_123', got '%s'", calls[0].ID)
	}

	var input map[string]string
	if err := json.Unmarshal(calls[0].Input, &input); err != nil {
		t.Fatalf("failed to unmarshal input: %v", err)
	}
	if input["param"] != "value" {
		t.Errorf("expected input param 'value', got '%s'", input["param"])
	}

	if !strings.Contains(remaining, "Here is some text before") {
		t.Errorf("remaining text should contain prefix, got: %s", remaining)
	}
	if !strings.Contains(remaining, "And some text after") {
		t.Errorf("remaining text should contain suffix, got: %s", remaining)
	}
	if strings.Contains(remaining, "<tool_call>") {
		t.Errorf("remaining text should not contain tool_call tags")
	}
}

func TestParseToolCalls_MultipleToolCalls(t *testing.T) {
	text := `<tool_call>
{"name": "tool1", "input": {"a": 1}}
</tool_call>
<tool_call>
{"name": "tool2", "input": {"b": 2}}
</tool_call>
<tool_call>
{"name": "tool3", "input": {"c": 3}}
</tool_call>`

	calls, remaining := ParseToolCalls(text)

	if len(calls) != 3 {
		t.Fatalf("expected 3 tool calls, got %d", len(calls))
	}

	expectedNames := []string{"tool1", "tool2", "tool3"}
	for i, name := range expectedNames {
		if calls[i].Name != name {
			t.Errorf("call %d: expected name '%s', got '%s'", i, name, calls[i].Name)
		}
	}

	// 每个调用应该有自动生成的 ID
	for i, call := range calls {
		if call.ID == "" {
			t.Errorf("call %d: expected auto-generated ID, got empty", i)
		}
		if !strings.HasPrefix(call.ID, "toolu_") {
			t.Errorf("call %d: expected ID to start with 'toolu_', got '%s'", i, call.ID)
		}
	}

	remaining = strings.TrimSpace(remaining)
	if remaining != "" {
		t.Errorf("expected empty remaining text, got: '%s'", remaining)
	}
}

func TestParseToolCalls_NoToolCalls(t *testing.T) {
	text := "This is just regular text without any tool calls."

	calls, remaining := ParseToolCalls(text)

	if len(calls) != 0 {
		t.Fatalf("expected 0 tool calls, got %d", len(calls))
	}

	if remaining != text {
		t.Errorf("expected remaining text to be unchanged, got: '%s'", remaining)
	}
}

func TestParseToolCalls_EmptyInput(t *testing.T) {
	text := `<tool_call>
{"name": "empty_tool"}
</tool_call>`

	calls, _ := ParseToolCalls(text)

	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}

	// 应该自动填充空的 input
	if string(calls[0].Input) != "{}" {
		t.Errorf("expected empty input '{}', got '%s'", string(calls[0].Input))
	}
}

func TestParseToolCalls_ComplexInput(t *testing.T) {
	text := `<tool_call>
{"name": "complex_tool", "input": {"nested": {"key": "value"}, "array": [1, 2, 3], "bool": true}}
</tool_call>`

	calls, _ := ParseToolCalls(text)

	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}

	var input map[string]interface{}
	if err := json.Unmarshal(calls[0].Input, &input); err != nil {
		t.Fatalf("failed to unmarshal complex input: %v", err)
	}

	if input["bool"] != true {
		t.Errorf("expected bool to be true")
	}

	array, ok := input["array"].([]interface{})
	if !ok || len(array) != 3 {
		t.Errorf("expected array with 3 elements")
	}

	nested, ok := input["nested"].(map[string]interface{})
	if !ok || nested["key"] != "value" {
		t.Errorf("expected nested object with key=value")
	}
}

func TestParseToolCalls_MalformedJSON(t *testing.T) {
	text := `<tool_call>
{"name": "valid_tool", "input": {}}
</tool_call>
<tool_call>
{invalid json here}
</tool_call>
<tool_call>
{"name": "another_valid", "input": {}}
</tool_call>`

	calls, remaining := ParseToolCalls(text)

	// 只有有效的 JSON 应该被解析
	if len(calls) != 2 {
		t.Fatalf("expected 2 valid tool calls, got %d", len(calls))
	}

	if calls[0].Name != "valid_tool" {
		t.Errorf("expected first call name 'valid_tool', got '%s'", calls[0].Name)
	}
	if calls[1].Name != "another_valid" {
		t.Errorf("expected second call name 'another_valid', got '%s'", calls[1].Name)
	}

	// 所有 tool_call 标签都应该被移除
	if strings.Contains(remaining, "<tool_call>") {
		t.Errorf("remaining should not contain tool_call tags")
	}
}

func TestParseToolCalls_MultilineInput(t *testing.T) {
	text := `<tool_call>
{
  "name": "multiline_tool",
  "input": {
    "code": "function test() {\n  return 42;\n}"
  }
}
</tool_call>`

	calls, _ := ParseToolCalls(text)

	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}

	if calls[0].Name != "multiline_tool" {
		t.Errorf("expected name 'multiline_tool', got '%s'", calls[0].Name)
	}
}

func TestParseToolCalls_WithExistingID(t *testing.T) {
	text := `<tool_call>
{"name": "tool_with_id", "id": "custom_id_123", "input": {}}
</tool_call>`

	calls, _ := ParseToolCalls(text)

	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}

	// 应该保留原有的 ID
	if calls[0].ID != "custom_id_123" {
		t.Errorf("expected ID 'custom_id_123', got '%s'", calls[0].ID)
	}
}

func TestParseToolCalls_WhitespaceHandling(t *testing.T) {
	text := `   <tool_call>
   {"name": "whitespace_tool", "input": {}}
   </tool_call>   `

	calls, remaining := ParseToolCalls(text)

	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}

	if calls[0].Name != "whitespace_tool" {
		t.Errorf("expected name 'whitespace_tool', got '%s'", calls[0].Name)
	}

	// remaining 应该被 trim
	if remaining != "" {
		t.Errorf("expected empty remaining after trim, got: '%s'", remaining)
	}
}
