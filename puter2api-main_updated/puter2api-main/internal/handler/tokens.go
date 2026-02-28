package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"puter2api/internal/parser"
	"puter2api/internal/storage"
	"puter2api/internal/types"

	"github.com/gin-gonic/gin"
)

// TokenHandler Token 管理处理器
type TokenHandler struct {
	storage *storage.Storage
}

// NewTokenHandler 创建 Token 处理器
func NewTokenHandler(s *storage.Storage) *TokenHandler {
	return &TokenHandler{storage: s}
}

// ListTokens 获取所有 Token
func (h *TokenHandler) ListTokens(c *gin.Context) {
	tokens, err := h.storage.GetAllTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 隐藏完整 token，只显示前后部分
	type TokenResponse struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Token     string `json:"token"` // 脱敏后的 token
		IsActive  bool   `json:"is_active"`
		IsValid   bool   `json:"is_valid"`
		LastUsed  string `json:"last_used,omitempty"`
		CreatedAt string `json:"created_at"`
	}

	var resp []TokenResponse
	for _, t := range tokens {
		masked := maskToken(t.Token)
		tr := TokenResponse{
			ID:        t.ID,
			Name:      t.Name,
			Token:     masked,
			IsActive:  t.IsActive,
			IsValid:   t.IsValid,
			CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if t.LastUsed != nil {
			tr.LastUsed = t.LastUsed.Format("2006-01-02 15:04:05")
		}
		resp = append(resp, tr)
	}

	c.JSON(http.StatusOK, gin.H{"tokens": resp})
}

// AddToken 添加新 Token
func (h *TokenHandler) AddToken(c *gin.Context) {
	var req struct {
		Name  string `json:"name"`
		Input string `json:"input"` // curl 命令或直接的 token
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 解析 token
	token, err := parser.ParseToken(req.Input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse token: " + err.Error()})
		return
	}

	// 添加到数据库
	t, err := h.storage.AddToken(req.Name, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "token added successfully",
		"id":      t.ID,
		"name":    t.Name,
		"token":   maskToken(t.Token),
	})
}

// DeleteToken 删除 Token
func (h *TokenHandler) DeleteToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.storage.DeleteToken(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token deleted"})
}

// UpdateTokenName 更新 Token 名称
func (h *TokenHandler) UpdateTokenName(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 获取现有 token
	t, err := h.storage.GetToken(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}

	// 只更新名称
	if err := h.storage.UpdateToken(id, req.Name, t.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token updated"})
}

// ToggleToken 切换 Token 启用状态
func (h *TokenHandler) ToggleToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.storage.UpdateTokenActive(id, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token updated"})
}

// TestToken 测试 Token 是否有效
func (h *TokenHandler) TestToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 获取 token
	t, err := h.storage.GetToken(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}

	// 测试 token
	isValid, testResult := testPuterToken(t.Token)

	// 更新有效性
	h.storage.UpdateTokenValid(id, isValid)

	c.JSON(http.StatusOK, gin.H{
		"is_valid": isValid,
		"message":  testResult,
	})
}

// TestAllTokens 测试所有 Token
func (h *TokenHandler) TestAllTokens(c *gin.Context) {
	tokens, err := h.storage.GetAllTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type Result struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		IsValid bool   `json:"is_valid"`
		Message string `json:"message"`
	}

	var results []Result
	for _, t := range tokens {
		isValid, msg := testPuterToken(t.Token)
		h.storage.UpdateTokenValid(t.ID, isValid)
		results = append(results, Result{
			ID:      t.ID,
			Name:    t.Name,
			IsValid: isValid,
			Message: msg,
		})
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// testPuterToken 测试 Puter token 是否有效
func testPuterToken(token string) (bool, string) {
	// 创建一个简单的测试请求
	client := NewPuterTestClient()

	// 发送一个简单的测试消息
	messages := []types.PuterMessage{
		{Role: "user", Content: "Hi"},
	}

	resp, err := client.TestToken(messages, token)
	if err != nil {
		return false, "Error: " + err.Error()
	}

	if resp != "" {
		return true, "Token is valid"
	}

	return false, "No response received"
}

// maskToken 脱敏 token，只显示前10和后10个字符
func maskToken(token string) string {
	if len(token) <= 24 {
		return token
	}
	return token[:10] + "..." + token[len(token)-10:]
}

// PuterTestClient 用于测试 token 的简单客户端
type PuterTestClient struct {
	httpClient *http.Client
}

// NewPuterTestClient 创建测试客户端
func NewPuterTestClient() *PuterTestClient {
	return &PuterTestClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// TestToken 测试 token 是否有效
func (c *PuterTestClient) TestToken(messages []types.PuterMessage, authToken string) (string, error) {
	puterReq := types.PuterRequest{
		Interface: "puter-chat-completion",
		Driver:    "claude",
		TestMode:  false,
		Method:    "complete",
		Args: types.PuterArgs{
			Messages: messages,
			Model:    "claude-sonnet-4-5-20250514",
			Stream:   false,
		},
		AuthToken: authToken,
	}

	body, _ := json.Marshal(puterReq)

	req, err := http.NewRequest("POST", "https://api.puter.com/drivers/call", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("content-type", "text/plain;actually=json")
	req.Header.Set("origin", "https://docs.puter.com")
	req.Header.Set("referer", "https://docs.puter.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return string(respBody), nil
}
