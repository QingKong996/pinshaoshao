package router

import (
	"pinshaoshao/internal/controller"

	"github.com/gin-gonic/gin"
)

func UserRegister(r *gin.Engine, u *controller.UserController) {
	api := r.Group("/api")
	v1 := api.Group("/v1")

	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1.POST("/login", u.Login)
}
