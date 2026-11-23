package service

import (
	"context"
	"fmt"

	"pinshaoshao/internal/feature/feed/apperr"
	"pinshaoshao/internal/feature/feed/model"
)

const (
	defaultFeedPageSize = 20
	maxFeedPageSize     = 50
	maxFeedPage         = 1_000_000
)

type FeedRepository interface {
	List(ctx context.Context, userID int, page int, pageSize int) (*model.FeedResult, error)
	Get(ctx context.Context, productID int) (*model.ProductDetail, error)
}

type FeedService struct {
	feedRepository FeedRepository
}

func NewFeedService(feedRepository FeedRepository) *FeedService {
	return &FeedService{feedRepository: feedRepository}
}

func (s *FeedService) List(ctx context.Context, userID int, page int, pageSize int) (*model.FeedResult, error) {
	normalized, err := normalizeFeedQuery(userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("service.ListFeed: %w", err)
	}

	result, err := s.feedRepository.List(ctx, normalized.userID, normalized.page, normalized.pageSize)
	if err != nil {
		return nil, fmt.Errorf("service.ListFeed: %w", err)
	}
	return result, nil
}

func (s *FeedService) Get(ctx context.Context, productID int) (*model.ProductDetail, error) {
	if productID <= 0 {
		return nil, fmt.Errorf("service.GetProductDetail: %w", apperr.ErrInvalidProductID)
	}

	product, err := s.feedRepository.Get(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("service.GetProductDetail: %w", err)
	}
	return product, nil
}

type feedQuery struct {
	userID   int
	page     int
	pageSize int
}

func normalizeFeedQuery(userID int, page int, pageSize int) (feedQuery, error) {
	if userID <= 0 {
		return feedQuery{}, apperr.ErrInvalidFeedQuery
	}
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = defaultFeedPageSize
	}
	if page < 1 || page > maxFeedPage || pageSize < 1 || pageSize > maxFeedPageSize {
		return feedQuery{}, apperr.ErrInvalidFeedQuery
	}
	return feedQuery{
		userID:   userID,
		page:     page,
		pageSize: pageSize,
	}, nil
}
