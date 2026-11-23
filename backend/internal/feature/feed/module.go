package feed

import (
	feedcontroller "pinshaoshao/internal/feature/feed/controller"
	feedrepository "pinshaoshao/internal/feature/feed/repository"
	feedservice "pinshaoshao/internal/feature/feed/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	controller *feedcontroller.FeedController
}

func NewModule(pool *pgxpool.Pool) *Module {
	feedRepository := feedrepository.NewFeedRepository(pool)
	feedService := feedservice.NewFeedService(feedRepository)
	return &Module{
		controller: feedcontroller.NewFeedController(feedService),
	}
}

func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	protected := r.Group("/api/v1")
	protected.Use(authMiddleware)
	protected.GET("/feed", m.controller.List)
}
