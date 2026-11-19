package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// pathID parses a positive integer route parameter.
func pathID(ctx *gin.Context, name string) (int, bool) {
	value, err := strconv.Atoi(ctx.Param(name))
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}
