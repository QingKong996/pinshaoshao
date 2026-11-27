package middleware

import (
	"bytes"
	"io"
	"net/http"

	"pinshaoshao/internal/http/httpx"

	"github.com/gin-gonic/gin"
)

const (
	DefaultRequestBodyLimit  int64 = 1 << 20
	requestBodyTooLargeError       = "request body too large"
)

func RequestBodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Body == http.NoBody {
			c.Next()
			return
		}
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, httpx.ErrorResponse{
				Error: requestBodyTooLargeError,
			})
			return
		}

		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBytes+1))
		_ = c.Request.Body.Close()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, httpx.ErrorResponse{
				Error: httpx.InvalidJSONMessage,
			})
			return
		}
		if int64(len(body)) > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, httpx.ErrorResponse{
				Error: requestBodyTooLargeError,
			})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}
