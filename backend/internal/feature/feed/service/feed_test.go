package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"pinshaoshao/internal/feature/feed/apperr"
	"pinshaoshao/internal/feature/feed/model"
	"pinshaoshao/internal/feature/feed/service"
)

type stubFeedRepository struct {
	calls     int
	userID    int
	page      int
	pageSize  int
	productID int
	result    *model.FeedResult
	detail    *model.ProductDetail
	err       error
}

func (r *stubFeedRepository) List(
	_ context.Context,
	userID int,
	page int,
	pageSize int,
) (*model.FeedResult, error) {
	r.calls++
	r.userID = userID
	r.page = page
	r.pageSize = pageSize
	return r.result, r.err
}

func (r *stubFeedRepository) Get(
	_ context.Context,
	productID int,
) (*model.ProductDetail, error) {
	r.calls++
	r.productID = productID
	return r.detail, r.err
}

func TestFeedServiceListNormalizesPagination(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{name: "defaults", wantPage: 1, wantPageSize: 20},
		{name: "explicit values", page: 3, pageSize: 50, wantPage: 3, wantPageSize: 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &model.FeedResult{
				Items:    []model.FeedItem{{ID: 1}},
				Page:     tt.wantPage,
				PageSize: tt.wantPageSize,
				Total:    1,
			}
			repository := &stubFeedRepository{result: result}

			got, err := service.NewFeedService(repository).List(
				context.Background(),
				7,
				tt.page,
				tt.pageSize,
			)
			if err != nil {
				t.Fatalf("list feed: %v", err)
			}
			if got != result {
				t.Fatalf("result = %p, want %p", got, result)
			}
			if repository.userID != 7 ||
				repository.page != tt.wantPage ||
				repository.pageSize != tt.wantPageSize {
				t.Fatalf(
					"repository args = %d/%d/%d, want 7/%d/%d",
					repository.userID,
					repository.page,
					repository.pageSize,
					tt.wantPage,
					tt.wantPageSize,
				)
			}
		})
	}
}

func TestFeedServiceListRejectsInvalidQuery(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		page     int
		pageSize int
	}{
		{name: "missing user", page: 1, pageSize: 20},
		{name: "negative page", userID: 7, page: -1, pageSize: 20},
		{name: "page too large", userID: 7, page: 1_000_001, pageSize: 20},
		{name: "negative page size", userID: 7, page: 1, pageSize: -1},
		{name: "page size too large", userID: 7, page: 1, pageSize: 51},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &stubFeedRepository{}

			_, err := service.NewFeedService(repository).List(
				context.Background(),
				tt.userID,
				tt.page,
				tt.pageSize,
			)
			if !errors.Is(err, apperr.ErrInvalidFeedQuery) {
				t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidFeedQuery)
			}
			if repository.calls != 0 {
				t.Fatalf("repository calls = %d, want 0", repository.calls)
			}
		})
	}
}

func TestFeedServiceListPreservesRepositoryErrors(t *testing.T) {
	repositoryErr := errors.New("database is down")
	repository := &stubFeedRepository{err: repositoryErr}

	_, err := service.NewFeedService(repository).List(
		context.Background(),
		7,
		1,
		20,
	)
	if !errors.Is(err, repositoryErr) {
		t.Fatalf("error = %v, want %v", err, repositoryErr)
	}
}

func TestFeedServiceGet(t *testing.T) {
	detail := &model.ProductDetail{ID: 10}
	repository := &stubFeedRepository{detail: detail}

	got, err := service.NewFeedService(repository).Get(context.Background(), 10)
	if err != nil {
		t.Fatalf("get product detail: %v", err)
	}
	if got != detail {
		t.Fatalf("product detail = %p, want %p", got, detail)
	}
	if repository.productID != 10 {
		t.Fatalf("product id = %d, want 10", repository.productID)
	}
}

func TestFeedServiceGetRejectsInvalidID(t *testing.T) {
	for _, productID := range []int{0, -1} {
		t.Run(fmt.Sprintf("product %d", productID), func(t *testing.T) {
			repository := &stubFeedRepository{}

			_, err := service.NewFeedService(repository).Get(context.Background(), productID)
			if !errors.Is(err, apperr.ErrInvalidProductID) {
				t.Fatalf("error = %v, want %v", err, apperr.ErrInvalidProductID)
			}
			if repository.calls != 0 {
				t.Fatalf("repository calls = %d, want 0", repository.calls)
			}
		})
	}
}

func TestFeedServiceGetPreservesRepositoryErrors(t *testing.T) {
	repositoryErr := errors.New("database is down")
	repository := &stubFeedRepository{err: repositoryErr}

	_, err := service.NewFeedService(repository).Get(context.Background(), 10)
	if !errors.Is(err, repositoryErr) {
		t.Fatalf("error = %v, want %v", err, repositoryErr)
	}
}
