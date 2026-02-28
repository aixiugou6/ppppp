package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"puter2api/internal/handler"
	"puter2api/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed web/*
var webFS embed.FS

func main() {
	// 配置 zerolog 彩色控制台输出
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./puter2api.db"
	}

	// 初始化数据库
	store, err := storage.New(dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化数据库失败")
	}
	defer store.Close()

	// 创建处理器 - 从数据库获取 Token
	h := handler.NewHandler(store)
	th := handler.NewTokenHandler(store)

	// 设置 Gin 使用 zerolog
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginLogger())

	// Claude API 兼容端点
	r.POST("/v1/messages", h.HandleMessages)
	r.POST("/messages", h.HandleMessages)

	// OpenAI API 兼容端点
	r.POST("/v1/chat/completions", h.HandleOpenAIChat)
	r.GET("/v1/models", h.HandleModels)

	// Token 管理 API
	api := r.Group("/api")
	{
		api.GET("/tokens", th.ListTokens)
		api.POST("/tokens", th.AddToken)
		api.DELETE("/tokens/:id", th.DeleteToken)
		api.PUT("/tokens/:id", th.UpdateTokenName)
		api.PUT("/tokens/:id/toggle", th.ToggleToken)
		api.POST("/tokens/:id/test", th.TestToken)
		api.POST("/tokens/test-all", th.TestAllTokens)
	}

	// 静态文件服务 (Web UI)
	webContent, _ := fs.Sub(webFS, "web")
	r.StaticFS("/ui", http.FS(webContent))

	// 根路径重定向到 UI
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui/")
	})

	// 启动服务
	log.Info().Str("port", port).Msg("服务启动")
	log.Info().Msgf("Web UI: http://localhost:%s/ui/", port)
	r.Run(":" + port)
}

// ginLogger 自定义 Gin 日志中间件，使用 zerolog
func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		// 根据状态码选择日志级别
		var event *zerolog.Event
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		} else {
			event = log.Info()
		}

		event.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Str("耗时", fmt.Sprintf("%.2fs", latency)).
			Msg("GIN")
	}
}
