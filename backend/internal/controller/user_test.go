package controller_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/security"
	"pinshaoshao/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

const testUsername = "alice01"

func newTestUserController(t *testing.T) (*controller.UserController, *mocks.MockUserRepository) {
	t.Helper()

	userRepository := mocks.NewMockUserRepository(gomock.NewController(t))
	return controller.NewUserController(service.NewUserService(userRepository)), userRepository
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
		userController, userRepository := newTestUserController(t)
		user := &model.User{
			ID:           1,
			Username:     testUsername,
			PasswordHash: mustHashPassword(t, testPassword),
			Status:       model.UserStatusActive,
		}
		userRepository.EXPECT().
			GetUserByUsername(gomock.Any(), testUsername).
			Return(user, nil)
		userRepository.EXPECT().
			CreateSession(gomock.Any(), user, gomock.Any()).
			Return(nil)

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		userController.Login(ctx)

		requireStatus(t, rec, http.StatusOK)
		requireToken(t, rec)
	})

	t.Run("rejects invalid credentials", func(t *testing.T) {
		userController, userRepository := newTestUserController(t)
		user := &model.User{
			ID:           1,
			Username:     testUsername,
			PasswordHash: mustHashPassword(t, "OtherPassword123"),
			Status:       model.UserStatusActive,
		}
		userRepository.EXPECT().
			GetUserByUsername(gomock.Any(), testUsername).
			Return(user, nil)

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
		userController, userRepository := newTestUserController(t)
		user := &model.User{
			ID:           1,
			Username:     testUsername,
			PasswordHash: mustHashPassword(t, testPassword),
			Status:       model.UserStatusBanned,
		}
		userRepository.EXPECT().
			GetUserByUsername(gomock.Any(), testUsername).
			Return(user, nil)

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		userController.Login(ctx)

		requireStatus(t, rec, http.StatusForbidden)
		requireError(t, rec, apperr.ErrUserBanned)
	})

	t.Run("returns 500 on repository failure", func(t *testing.T) {
		userController, userRepository := newTestUserController(t)
		userRepository.EXPECT().
			GetUserByUsername(gomock.Any(), testUsername).
			Return(nil, errors.New("database is down"))

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		userController.Login(ctx)

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestUserRegister(t *testing.T) {
	registration := fmt.Sprintf(`{"username":%q,"password":%q}`, testUsername, testPassword)

	t.Run("returns a session token", func(t *testing.T) {
		userController, userRepository := newTestUserController(t)
		userRepository.EXPECT().
			CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, user *model.User, _ []byte) error {
				user.ID = 1
				return nil
			})

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", registration)
		userController.Register(ctx)

		requireStatus(t, rec, http.StatusCreated)
		requireToken(t, rec)
	})

	t.Run("rejects a taken username", func(t *testing.T) {
		userController, userRepository := newTestUserController(t)
		userRepository.EXPECT().
			CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("repository: %w", apperr.ErrUsernameTaken))

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", registration)
		userController.Register(ctx)

		requireStatus(t, rec, http.StatusConflict)
		requireError(t, rec, apperr.ErrUsernameTaken)
	})

	t.Run("rejects a taken email", func(t *testing.T) {
		userController, userRepository := newTestUserController(t)
		userRepository.EXPECT().
			CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("repository: %w", apperr.ErrEmailTaken))

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

	t.Run("returns 500 on repository failure", func(t *testing.T) {
		userController, userRepository := newTestUserController(t)
		userRepository.EXPECT().
			CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("database is down"))

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", registration)
		userController.Register(ctx)

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestUserLogout(t *testing.T) {
	activeUser := &model.User{ID: 1, Username: testUsername, Status: model.UserStatusActive}

	t.Run("revokes the session", func(t *testing.T) {
		userController, userRepository := newTestUserController(t)
		userRepository.EXPECT().
			DeleteSessionByTokenHash(gomock.Any(), security.HashSessionToken("session-token")).
			Return(nil)

		rec := doRequest(newUserLogoutRouter(t, userController, activeUser), http.MethodPost, "/logout", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusNoContent)
		if rec.Body.Len() != 0 {
			t.Fatalf("body = %q, want empty", rec.Body.String())
		}
	})

	t.Run("returns 500 when the session cannot be revoked", func(t *testing.T) {
		userController, userRepository := newTestUserController(t)
		userRepository.EXPECT().
			DeleteSessionByTokenHash(gomock.Any(), security.HashSessionToken("session-token")).
			Return(errors.New("database is down"))

		rec := doRequest(newUserLogoutRouter(t, userController, activeUser), http.MethodPost, "/logout", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}
