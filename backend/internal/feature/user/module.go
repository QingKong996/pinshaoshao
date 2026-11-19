package user

import (
	usercontroller "pinshaoshao/internal/feature/user/controller"
	userrepository "pinshaoshao/internal/feature/user/repository"
	userservice "pinshaoshao/internal/feature/user/service"
	"pinshaoshao/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	controller *usercontroller.UserController
	service    *userservice.UserService
}

func NewModule(pool *pgxpool.Pool) *Module {
	userRepository := userrepository.NewUserRepository(pool)
	userService := userservice.NewUserService(userRepository)
	return &Module{
		controller: usercontroller.NewUserController(userService),
		service:    userService,
	}
}

func (m *Module) AuthMiddleware() gin.HandlerFunc {
	return middleware.UserAuth(m.service)
}

func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	v1 := r.Group("/api/v1")

	v1.POST("/login", m.controller.Login)
	v1.POST("/register", m.controller.Register)

	protected := v1.Group("")
	protected.Use(authMiddleware)
	protected.POST("/logout", m.controller.Logout)
}
