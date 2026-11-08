package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/security"
	"pinshaoshao/internal/service"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

const (
	testUsername = "alice"
	testPassword = "Password123"
)

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return hash
}

func newTestUserRouter(t *testing.T) (http.Handler, *mocks.MockUserRepository) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	userRepository := mocks.NewMockUserRepository(gomock.NewController(t))
	userService := service.NewUserService(userRepository)
	userController := controller.NewUserController(userService)

	r := gin.New()
	UserRegister(r, userController, middleware.UserAuth(userService))

	// Test-only route to observe the user injected by UserAuth.
	r.GET("/api/v1/whoami", middleware.UserAuth(userService), func(c *gin.Context) {
		user, ok := middleware.CurrentUser(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user missing from context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"username": user.Username})
	})

	return r, userRepository
}

func doRequest(r http.Handler, method, path, authorization, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func requireStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
	}
}

func responseToken(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	if body.Token == "" {
		t.Fatalf("response has no token: %s", rec.Body.String())
	}
	return body.Token
}

func responseError(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	return body.Error
}

func TestUserRoutes(t *testing.T) {
	r, userRepository := newTestUserRouter(t)
	credentials := `{"username":"` + testUsername + `","password":"` + testPassword + `"}`
	passwordHash := mustHashPassword(t, testPassword)

	activeUser := &model.User{
		ID:           1,
		Username:     testUsername,
		PasswordHash: passwordHash,
		Status:       model.UserStatusActive,
	}
	bannedUser := &model.User{
		ID:           1,
		Username:     testUsername,
		PasswordHash: passwordHash,
		Status:       model.UserStatusBanned,
	}

	var token string

	t.Run("register issues a session token", func(t *testing.T) {
		userRepository.EXPECT().
			CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, user *model.User, _ []byte) error {
				user.ID = 1
				user.Status = model.UserStatusActive
				return nil
			})

		rec := doRequest(r, http.MethodPost, "/api/v1/register", "", credentials)
		requireStatus(t, rec, http.StatusCreated)
		token = responseToken(t, rec)
	})

	t.Run("middleware rejects missing credentials", func(t *testing.T) {
		rec := doRequest(r, http.MethodPost, "/api/v1/logout", "", "")
		requireStatus(t, rec, http.StatusUnauthorized)
		if got := responseError(t, rec); got != apperr.ErrUnauthorized.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrUnauthorized.Error())
		}
		if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
			t.Fatalf("WWW-Authenticate = %q, want %q", got, "Bearer")
		}
	})

	t.Run("middleware rejects non-bearer scheme", func(t *testing.T) {
		requireStatus(t, doRequest(r, http.MethodGet, "/api/v1/whoami", "Basic abc", ""), http.StatusUnauthorized)
	})

	t.Run("middleware rejects unknown token", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken("not-a-real-token")).
			Return(nil, fmt.Errorf("repository.GetUserBySessionTokenHash: %w", apperr.ErrInvalidSession))

		rec := doRequest(r, http.MethodGet, "/api/v1/whoami", "Bearer not-a-real-token", "")
		requireStatus(t, rec, http.StatusUnauthorized)
		if got := responseError(t, rec); got != apperr.ErrInvalidSession.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrInvalidSession.Error())
		}
	})

	t.Run("middleware injects user for valid token", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeUser, nil)

		rec := doRequest(r, http.MethodGet, "/api/v1/whoami", "Bearer "+token, "")
		requireStatus(t, rec, http.StatusOK)

		var whoami struct {
			Username string `json:"username"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &whoami); err != nil {
			t.Fatalf("decode whoami response: %v", err)
		}
		if whoami.Username != testUsername {
			t.Fatalf("username = %q, want %q", whoami.Username, testUsername)
		}
	})

	t.Run("logout revokes the session token", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeUser, nil)
		userRepository.EXPECT().
			DeleteSessionByTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(nil)
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(nil, fmt.Errorf("repository.GetUserBySessionTokenHash: %w", apperr.ErrInvalidSession))

		requireStatus(t, doRequest(r, http.MethodPost, "/api/v1/logout", "Bearer "+token, ""), http.StatusNoContent)

		rec := doRequest(r, http.MethodGet, "/api/v1/whoami", "Bearer "+token, "")
		requireStatus(t, rec, http.StatusUnauthorized)
		if got := responseError(t, rec); got != apperr.ErrInvalidSession.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrInvalidSession.Error())
		}
	})

	t.Run("login issues a new session token", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserByUsername(gomock.Any(), testUsername).
			Return(activeUser, nil)
		userRepository.EXPECT().
			CreateSession(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		rec := doRequest(r, http.MethodPost, "/api/v1/login", "", credentials)
		requireStatus(t, rec, http.StatusOK)
		token = responseToken(t, rec)

		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeUser, nil)
		requireStatus(t, doRequest(r, http.MethodGet, "/api/v1/whoami", "Bearer "+token, ""), http.StatusOK)
	})

	t.Run("middleware rejects banned user", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(bannedUser, nil).
			Times(2)

		rec := doRequest(r, http.MethodGet, "/api/v1/whoami", "Bearer "+token, "")
		requireStatus(t, rec, http.StatusForbidden)
		if got := responseError(t, rec); got != apperr.ErrUserBanned.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrUserBanned.Error())
		}
		requireStatus(t, doRequest(r, http.MethodPost, "/api/v1/logout", "Bearer "+token, ""), http.StatusForbidden)
	})
}
