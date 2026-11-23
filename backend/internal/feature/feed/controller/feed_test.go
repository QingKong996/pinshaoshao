package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pinshaoshao/internal/feature/feed/apperr"
	"pinshaoshao/internal/feature/feed/controller"
	"pinshaoshao/internal/feature/feed/dto"
	"pinshaoshao/internal/feature/feed/model"
	usermodule "pinshaoshao/internal/feature/user"
	usermodel "pinshaoshao/internal/feature/user/model"
	"pinshaoshao/internal/http/httpx"

	"github.com/gin-gonic/gin"
)

type stubFeedService struct {
	calls    int
	userID   int
	page     int
	pageSize int
	result   *model.FeedResult
	err      error
}

func (s *stubFeedService) List(
	_ context.Context,
	userID int,
	page int,
	pageSize int,
) (*model.FeedResult, error) {
	s.calls++
	s.userID = userID
	s.page = page
	s.pageSize = pageSize
	return s.result, s.err
}

type stubFeedUserAuthenticator struct {
	user *usermodel.User
}

func (s stubFeedUserAuthenticator) Authenticate(
	context.Context,
	string,
) (*usermodel.User, error) {
	return s.user, nil
}

func newFeedRouter(feedController *controller.FeedController) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(usermodule.AuthMiddleware(stubFeedUserAuthenticator{
		user: &usermodel.User{
			ID:       7,
			Username: "alice01",
			Status:   usermodel.UserStatusActive,
		},
	}))
	protected.GET("/feed", feedController.List)
	return r
}

func TestFeedList(t *testing.T) {
	categoryName := "Food"
	feedService := &stubFeedService{
		result: &model.FeedResult{
			Items: []model.FeedItem{{
				ID:             1,
				ShopID:         2,
				ShopName:       "Example shop",
				Name:           "Coffee",
				Description:    "Roasted beans",
				CategoryName:   &categoryName,
				MinPriceCents:  1999,
				MaxPriceCents:  2999,
				AvailableStock: 8,
			}},
			Page:     2,
			PageSize: 10,
			Total:    25,
		},
	}
	feedController := controller.NewFeedController(feedService)

	rec := doFeedRequest(
		newFeedRouter(feedController),
		http.MethodGet,
		"/feed?page=2&page_size=10",
	)
	requireFeedStatus(t, rec, http.StatusOK)
	if feedService.userID != 7 || feedService.page != 2 || feedService.pageSize != 10 {
		t.Fatalf(
			"service args = %d/%d/%d, want 7/2/10",
			feedService.userID,
			feedService.page,
			feedService.pageSize,
		)
	}

	var body feeddto.FeedListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode feed response: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].Name != "Coffee" {
		t.Fatalf("items = %+v", body.Items)
	}
	if body.Items[0].AvailableStock != 8 ||
		body.Items[0].MinPriceCents != 1999 ||
		body.Items[0].MaxPriceCents != 2999 {
		t.Fatalf("item = %+v", body.Items[0])
	}
	if body.Page != 2 || body.PageSize != 10 || body.Total != 25 || body.TotalPages != 3 {
		t.Fatalf("pagination = %+v", body)
	}
}

func TestFeedListRejectsInvalidQuery(t *testing.T) {
	tests := []string{
		"/feed?page=0",
		"/feed?page=1000001",
		"/feed?page_size=0",
		"/feed?page_size=51",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			feedService := &stubFeedService{}
			feedController := controller.NewFeedController(feedService)

			rec := doFeedRequest(newFeedRouter(feedController), http.MethodGet, path)

			requireFeedStatus(t, rec, http.StatusBadRequest)
			requireFeedError(t, rec, apperr.ErrInvalidFeedQuery)
			if feedService.calls != 0 {
				t.Fatalf("service calls = %d, want 0", feedService.calls)
			}
		})
	}
}

func TestFeedListMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  error
	}{
		{
			name:       "invalid query",
			err:        apperr.ErrInvalidFeedQuery,
			wantStatus: http.StatusBadRequest,
			wantError:  apperr.ErrInvalidFeedQuery,
		},
		{
			name:       "service failure",
			err:        errors.New("database is down"),
			wantStatus: http.StatusInternalServerError,
			wantError:  apperr.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedService := &stubFeedService{err: tt.err}
			feedController := controller.NewFeedController(feedService)

			rec := doFeedRequest(newFeedRouter(feedController), http.MethodGet, "/feed")

			requireFeedStatus(t, rec, tt.wantStatus)
			requireFeedError(t, rec, tt.wantError)
		})
	}
}

func doFeedRequest(r http.Handler, method string, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer session-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func requireFeedStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
	}
}

func requireFeedError(t *testing.T, rec *httptest.ResponseRecorder, want error) {
	t.Helper()

	var body httpx.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode feed error: %v", err)
	}
	if body.Error != want.Error() {
		t.Fatalf("error = %q, want %q", body.Error, want.Error())
	}
}
