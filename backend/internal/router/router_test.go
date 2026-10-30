package router

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/service"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const (
	testUsername = "alice"
	testPassword = "Password123"
)

type fakeUserRepository struct {
	usersByUsername map[string]*model.User
	usersByID       map[int]*model.User
	sessions        map[string]int
	nextID          int
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		usersByUsername: make(map[string]*model.User),
		usersByID:       make(map[int]*model.User),
		sessions:        make(map[string]int),
	}
}

func (r *fakeUserRepository) GetUserByUsername(_ context.Context, username string) (*model.User, error) {
	user, ok := r.usersByUsername[username]
	if !ok {
		return nil, fmt.Errorf("fakeUserRepository.GetUserByUsername: %w", apperr.ErrUserNotFound)
	}
	return user, nil
}

func (r *fakeUserRepository) CreateUser(_ context.Context, user *model.User) error {
	if _, ok := r.usersByUsername[user.Username]; ok {
		return fmt.Errorf("fakeUserRepository.CreateUser: %w", apperr.ErrUsernameTaken)
	}
	if user.Status == "" {
		user.Status = model.UserStatusActive
	}
	r.nextID++
	user.ID = r.nextID
	r.usersByUsername[user.Username] = user
	r.usersByID[user.ID] = user
	return nil
}

func (r *fakeUserRepository) CreateSession(_ context.Context, user *model.User, tokenHash []byte) error {
	r.sessions[hex.EncodeToString(tokenHash)] = user.ID
	return nil
}

func (r *fakeUserRepository) GetUserBySessionTokenHash(_ context.Context, tokenHash []byte) (*model.User, error) {
	userID, ok := r.sessions[hex.EncodeToString(tokenHash)]
	if !ok {
		return nil, fmt.Errorf("fakeUserRepository.GetUserBySessionTokenHash: %w", apperr.ErrInvalidSession)
	}
	user, ok := r.usersByID[userID]
	if !ok {
		return nil, fmt.Errorf("fakeUserRepository.GetUserBySessionTokenHash: %w", apperr.ErrInvalidSession)
	}
	return user, nil
}

func (r *fakeUserRepository) DeleteSessionByTokenHash(_ context.Context, tokenHash []byte) error {
	delete(r.sessions, hex.EncodeToString(tokenHash))
	return nil
}

func (r *fakeUserRepository) SetUserStatus(username string, status model.UserStatus) {
	r.usersByUsername[username].Status = status
}

func newTestRouter(t *testing.T) (http.Handler, *fakeUserRepository) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	userRepository := newFakeUserRepository()
	userService := service.NewUserService(userRepository)
	userController := controller.NewUserController(userService)

	r := gin.New()
	UserRegister(r, userController, middleware.Auth(userService))

	// Test-only route to observe the user injected by Auth.
	r.GET("/api/v1/whoami", middleware.Auth(userService), func(c *gin.Context) {
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
	r, userRepository := newTestRouter(t)
	credentials := `{"username":"` + testUsername + `","password":"` + testPassword + `"}`

	var token string

	t.Run("public ping", func(t *testing.T) {
		requireStatus(t, doRequest(r, http.MethodGet, "/api/v1/ping", "", ""), http.StatusOK)
	})

	t.Run("register issues a session token", func(t *testing.T) {
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
		rec := doRequest(r, http.MethodGet, "/api/v1/whoami", "Bearer not-a-real-token", "")
		requireStatus(t, rec, http.StatusUnauthorized)
		if got := responseError(t, rec); got != apperr.ErrInvalidSession.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrInvalidSession.Error())
		}
	})

	t.Run("middleware injects user for valid token", func(t *testing.T) {
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
		requireStatus(t, doRequest(r, http.MethodPost, "/api/v1/logout", "Bearer "+token, ""), http.StatusNoContent)

		rec := doRequest(r, http.MethodGet, "/api/v1/whoami", "Bearer "+token, "")
		requireStatus(t, rec, http.StatusUnauthorized)
		if got := responseError(t, rec); got != apperr.ErrInvalidSession.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrInvalidSession.Error())
		}
	})

	t.Run("login issues a new session token", func(t *testing.T) {
		rec := doRequest(r, http.MethodPost, "/api/v1/login", "", credentials)
		requireStatus(t, rec, http.StatusOK)
		token = responseToken(t, rec)
		requireStatus(t, doRequest(r, http.MethodGet, "/api/v1/whoami", "Bearer "+token, ""), http.StatusOK)
	})

	t.Run("middleware rejects banned user", func(t *testing.T) {
		userRepository.SetUserStatus(testUsername, model.UserStatusBanned)

		rec := doRequest(r, http.MethodGet, "/api/v1/whoami", "Bearer "+token, "")
		requireStatus(t, rec, http.StatusForbidden)
		if got := responseError(t, rec); got != apperr.ErrUserBanned.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrUserBanned.Error())
		}
		requireStatus(t, doRequest(r, http.MethodPost, "/api/v1/logout", "Bearer "+token, ""), http.StatusForbidden)
	})
}
