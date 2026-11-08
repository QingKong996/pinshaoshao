package controller_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

const testUsername = "alice01"

func newTestUserController(t *testing.T) (*controller.UserController, *mocks.MockUserService) {
	t.Helper()

	userService := mocks.NewMockUserService(gomock.NewController(t))
	return controller.NewUserController(userService), userService
}

func newUserLogoutRouter(t *testing.T, userController *controller.UserController, user *model.User) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/logout", middleware.UserAuth(stubUserAuthenticator{user: user}), userController.Logout)
	return r
}

func TestUserLogin(t *testing.T) {
	credentials := fmt.Sprintf(`{"username":%q,"password":%q}`, testUsername, testPassword)

	t.Run("returns a session token", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return("session-token", nil)

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		userController.Login(ctx)

		requireStatus(t, rec, http.StatusOK)
		requireToken(t, rec)
	})

	t.Run("rejects invalid credentials", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("service: %w", apperr.ErrInvalidCredentials))

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		userController.Login(ctx)

		requireStatus(t, rec, http.StatusUnauthorized)
		requireError(t, rec, apperr.ErrInvalidCredentials)
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		userController, _ := newTestUserController(t)

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", `{bad`)
		userController.Login(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects an invalid payload", func(t *testing.T) {
		userController, _ := newTestUserController(t)

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", `{"username":"al"}`)
		userController.Login(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects a banned user", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("service: %w", apperr.ErrUserBanned))

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		userController.Login(ctx)

		requireStatus(t, rec, http.StatusForbidden)
		requireError(t, rec, apperr.ErrUserBanned)
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return("", errors.New("database is down"))

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		userController.Login(ctx)

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestUserRegister(t *testing.T) {
	registration := fmt.Sprintf(`{"username":%q,"password":%q}`, testUsername, testPassword)

	t.Run("returns a session token", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return("session-token", nil)

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", registration)
		userController.Register(ctx)

		requireStatus(t, rec, http.StatusCreated)
		requireToken(t, rec)
	})

	t.Run("rejects a taken username", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("service: %w", apperr.ErrUsernameTaken))

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", registration)
		userController.Register(ctx)

		requireStatus(t, rec, http.StatusConflict)
		requireError(t, rec, apperr.ErrUsernameTaken)
	})

	t.Run("rejects a taken email", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("service: %w", apperr.ErrEmailTaken))

		body := fmt.Sprintf(`{"username":%q,"password":%q,"email":"%s@example.com"}`, testUsername, testPassword, testUsername)
		ctx, rec := newControllerContext(t, http.MethodPost, "/register", body)
		userController.Register(ctx)

		requireStatus(t, rec, http.StatusConflict)
		requireError(t, rec, apperr.ErrEmailTaken)
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		userController, _ := newTestUserController(t)

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", `{bad`)
		userController.Register(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects an invalid payload", func(t *testing.T) {
		userController, _ := newTestUserController(t)

		body := fmt.Sprintf(`{"username":%q,"password":"short"}`, testUsername)
		ctx, rec := newControllerContext(t, http.MethodPost, "/register", body)
		userController.Register(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return("", errors.New("database is down"))

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", registration)
		userController.Register(ctx)

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestUserLogout(t *testing.T) {
	activeUser := &model.User{ID: 1, Username: testUsername, Status: model.UserStatusActive}

	t.Run("revokes the session", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Logout(gomock.Any(), "session-token").
			Return(nil)

		rec := doRequest(newUserLogoutRouter(t, userController, activeUser), http.MethodPost, "/logout", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusNoContent)
		if rec.Body.Len() != 0 {
			t.Fatalf("body = %q, want empty", rec.Body.String())
		}
	})

	t.Run("returns 500 when the session cannot be revoked", func(t *testing.T) {
		userController, userService := newTestUserController(t)
		userService.EXPECT().
			Logout(gomock.Any(), "session-token").
			Return(errors.New("database is down"))

		rec := doRequest(newUserLogoutRouter(t, userController, activeUser), http.MethodPost, "/logout", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}
