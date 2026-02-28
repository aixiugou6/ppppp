package parser

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

// ParsedCurl 解析后的 curl 信息
type ParsedCurl struct {
	AuthToken string `json:"auth_token"`
	Source    string `json:"source"` // 来源: "cookie", "body", "header"
}

// ParseCurl 从 curl 命令中提取 Puter 认证信息
func ParseCurl(curlCmd string) (*ParsedCurl, error) {
	if curlCmd == "" {
		return nil, errors.New("empty curl command")
	}

	// 尝试多种方式提取 token
	result := &ParsedCurl{}

	// 1. 从 Cookie 中提取 puter_auth_token
	if token := extractFromCookie(curlCmd); token != "" {
		result.AuthToken = token
		result.Source = "cookie"
		return result, nil
	}

	// 2. 从请求体中提取 auth_token
	if token := extractFromBody(curlCmd); token != "" {
		result.AuthToken = token
		result.Source = "body"
		return result, nil
	}

	// 3. 从 Authorization header 中提取
	if token := extractFromAuthHeader(curlCmd); token != "" {
		result.AuthToken = token
		result.Source = "header"
		return result, nil
	}

	// 4. 直接查找 JWT token 模式
	if token := extractJWT(curlCmd); token != "" {
		result.AuthToken = token
		result.Source = "jwt"
		return result, nil
	}

	return nil, errors.New("no auth token found in curl command")
}

// extractFromCookie 从 Cookie header 中提取 puter_auth_token
func extractFromCookie(curlCmd string) string {
	// 匹配 -H 'Cookie: ...' 或 --cookie '...'
	patterns := []string{
		`(?i)-H\s+['"]Cookie:\s*([^'"]+)['"]`,
		`(?i)--cookie\s+['"]([^'"]+)['"]`,
		`(?i)-b\s+['"]([^'"]+)['"]`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(curlCmd)
		if len(matches) > 1 {
			cookieStr := matches[1]
			// 从 cookie 字符串中提取 puter_auth_token
			if token := extractTokenFromCookieString(cookieStr); token != "" {
				return token
			}
		}
	}
	return ""
}

// extractTokenFromCookieString 从 cookie 字符串中提取 puter_auth_token
func extractTokenFromCookieString(cookieStr string) string {
	// 匹配 puter_auth_token=xxx
	re := regexp.MustCompile(`puter_auth_token=([^;\s]+)`)
	matches := re.FindStringSubmatch(cookieStr)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractFromBody 从请求体中提取 auth_token
func extractFromBody(curlCmd string) string {
	// 匹配 -d '...' 或 --data '...' 或 --data-raw '...'
	patterns := []string{
		`(?i)-d\s+['"](.+?)['"]`,
		`(?i)--data\s+['"](.+?)['"]`,
		`(?i)--data-raw\s+['"](.+?)['"]`,
		`(?i)--data-binary\s+['"](.+?)['"]`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(curlCmd)
		if len(matches) > 1 {
			bodyStr := matches[1]
			// 尝试解析 JSON
			if token := extractTokenFromJSON(bodyStr); token != "" {
				return token
			}
		}
	}
	return ""
}

// extractTokenFromJSON 从 JSON 字符串中提取 auth_token
func extractTokenFromJSON(jsonStr string) string {
	// 先尝试直接解析 JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err == nil {
		if token, ok := data["auth_token"].(string); ok {
			return token
		}
	}

	// 如果 JSON 解析失败，使用正则匹配
	re := regexp.MustCompile(`"auth_token"\s*:\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(jsonStr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractFromAuthHeader 从 Authorization header 中提取 token
func extractFromAuthHeader(curlCmd string) string {
	// 匹配 -H 'Authorization: Bearer xxx'
	re := regexp.MustCompile(`(?i)-H\s+['"]Authorization:\s*Bearer\s+([^'"]+)['"]`)
	matches := re.FindStringSubmatch(curlCmd)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractJWT 直接从字符串中提取 JWT token
func extractJWT(curlCmd string) string {
	// JWT 格式: xxxxx.xxxxx.xxxxx (三段 base64)
	re := regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)
	matches := re.FindString(curlCmd)
	return matches
}

// ParseToken 直接解析 token 字符串（支持纯 token 或 curl 命令）
func ParseToken(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("empty input")
	}

	// 如果看起来像 curl 命令
	if strings.HasPrefix(strings.ToLower(input), "curl ") {
		parsed, err := ParseCurl(input)
		if err != nil {
			return "", err
		}
		return parsed.AuthToken, nil
	}

	// 如果看起来像 JWT token
	if strings.HasPrefix(input, "eyJ") && strings.Count(input, ".") == 2 {
		return input, nil
	}

	// 尝试从中提取 JWT
	if token := extractJWT(input); token != "" {
		return token, nil
	}

	return "", errors.New("invalid token format")
}
