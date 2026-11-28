package middleware

import (
	"log/slog"
	"net/http"
	"reflect"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(ctx *gin.Context, recovered any) {
		logger.ErrorContext(ctx.Request.Context(), "http panic recovered",
			"method", ctx.Request.Method,
			"route", ctx.FullPath(),
			"panic_type", reflect.TypeOf(recovered).String(),
			"stack", string(debug.Stack()),
		)
		ctx.AbortWithStatus(http.StatusInternalServerError)
	})
}
