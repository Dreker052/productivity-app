package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		msg := "HTTP Request"

		attr := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.Duration("latency", duration),
			slog.String("clientIP", clientIP),
			slog.String("user_agent", userAgent),
		}

		if status >= 500 {
			logger.LogAttrs(c, slog.LevelError, msg, attr...)
		} else if status >= 400 {
			logger.LogAttrs(c, slog.LevelWarn, msg, attr...)
		} else {
			logger.LogAttrs(c, slog.LevelInfo, msg, attr...)
		}
	}
}
