package user

import (
	"context"

	usercontroller "pinshaoshao/internal/feature/user/controller"
	usermodel "pinshaoshao/internal/feature/user/model"
	userrepository "pinshaoshao/internal/feature/user/repository"
	userservice "pinshaoshao/internal/feature/user/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionAuthenticator resolves user session tokens.
type SessionAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*usermodel.User, error)
}

// Module wires the user feature and owns its routes.
type Module struct {
	controller *usercontroller.UserController
	service    *userservice.UserService
}

// NewModule builds the user feature from its dependencies.
func NewModule(pool *pgxpool.Pool) *Module {
	userRepository := userrepository.NewUserRepository(pool)
	userService := userservice.NewUserService(userRepository)
	return &Module{
		controller: usercontroller.NewUserController(userService),
		service:    userService,
	}
}

// AuthMiddleware protects user routes with the module's session authenticator.
func (m *Module) AuthMiddleware() gin.HandlerFunc {
	return AuthMiddleware(m.service)
}

// AuthMiddlewareWithGuard protects user routes with authentication and a guard.
func (m *Module) AuthMiddlewareWithGuard(guard gin.HandlerFunc) gin.HandlerFunc {
	return AuthMiddlewareWithGuard(m.service, guard)
}

// AuthMiddlewareWithGuards protects user routes with authentication and guards.
func (m *Module) AuthMiddlewareWithGuards(
	guard gin.HandlerFunc,
	unauthenticatedGuard gin.HandlerFunc,
) gin.HandlerFunc {
	return AuthMiddlewareWithGuards(m.service, guard, unauthenticatedGuard)
}

// Register mounts the user routes on the engine.
func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	v1 := r.Group("/api/v1")

	v1.POST("/login", m.controller.Login)
	v1.POST("/register", m.controller.Register)

	protected := v1.Group("")
	protected.Use(authMiddleware)
	protected.POST("/logout", m.controller.Logout)
}
