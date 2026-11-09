package service

import (
	"context"
	"fmt"
	"pinshaoshao/internal/model"
)

// DashboardRepository is the persistence the dashboard service depends on.
type DashboardRepository interface {
	GetDashboardStats(ctx context.Context, shopID int) (*model.DashboardStats, error)
}

type DashboardService struct {
	dashboardRepository DashboardRepository
}

func NewDashboardService(dashboardRepository DashboardRepository) *DashboardService {
	return &DashboardService{dashboardRepository: dashboardRepository}
}

// Overview returns the shop-wide counters.
func (s *DashboardService) Overview(ctx context.Context, shopID int) (*model.DashboardStats, error) {
	stats, err := s.dashboardRepository.GetDashboardStats(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("service.GetDashboard: %w", err)
	}
	return stats, nil
}
