package router

import (
	"pinshaoshao/internal/controller"

	"github.com/gin-gonic/gin"
)

// UserRegister wires the user endpoints.
func UserRegister(r *gin.Engine, u *controller.UserController, authMiddleware gin.HandlerFunc) {
	api := r.Group("/api")
	v1 := api.Group("/v1")

	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1.POST("/login", u.Login)
	v1.POST("/register", u.Register)

	protected := v1.Group("")
	protected.Use(authMiddleware)
	protected.POST("/logout", u.Logout)
}
