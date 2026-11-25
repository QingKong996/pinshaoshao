package middleware

import "github.com/gin-gonic/gin"

const contentSecurityPolicy = "default-src 'none'; base-uri 'none'; " +
	"form-action 'self'; frame-ancestors 'none'; style-src 'unsafe-inline'"

// SecurityHeaders applies browser security defaults to every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", contentSecurityPolicy)
		c.Header("Permissions-Policy", "camera=(), geolocation=(), microphone=()")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	}
}
