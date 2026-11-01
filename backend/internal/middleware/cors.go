package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	corsAllowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowHeaders = "Authorization, Content-Type"
	corsMaxAge       = "600"
)

// CORS allows browser clients from the configured origins to call the API.
// Origins are comma-separated; a single "*" allows every origin.
func CORS(allowedOrigins string) gin.HandlerFunc {
	allowAll := false
	allowed := make(map[string]struct{})
	for _, origin := range strings.Split(allowedOrigins, ",") {
		origin = strings.TrimSpace(origin)
		switch origin {
		case "":
			continue
		case "*":
			allowAll = true
		default:
			allowed[origin] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if _, ok := allowed[origin]; ok {
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Next()
			return
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", corsAllowMethods)
			c.Header("Access-Control-Allow-Headers", corsAllowHeaders)
			c.Header("Access-Control-Max-Age", corsMaxAge)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
