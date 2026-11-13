package service_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/security"
	"pinshaoshao/internal/service"

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

func newTestUserService(t *testing.T) (*service.UserService, *mocks.MockUserRepository) {
	t.Helper()

	userRepository := mocks.NewMockUserRepository(gomock.NewController(t))
	return service.NewUserService(userRepository), userRepository
}

func requireErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want %v", err, target)
	}
}

func TestAuthenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("valid session", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)
		token := "session-token"
		user := &model.User{ID: 1, Username: testUsername, Status: model.UserStatusActive}

		userRepository.EXPECT().
			GetUserBySessionTokenHash(ctx, security.HashSessionToken(token)).
			Return(user, nil)

		got, err := userService.Authenticate(ctx, token)
		if err != nil {
			t.Fatalf("authenticate: %v", err)
		}
		if got.Username != testUsername {
			t.Fatalf("username = %q, want %q", got.Username, testUsername)
		}
	})

	t.Run("unknown token", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)
		token := "not-a-real-token"

		userRepository.EXPECT().
			GetUserBySessionTokenHash(ctx, security.HashSessionToken(token)).
			Return(nil, fmt.Errorf("user repository: %w", apperr.ErrInvalidSession))

		_, err := userService.Authenticate(ctx, token)
		requireErrorIs(t, err, apperr.ErrInvalidSession)
	})

	t.Run("empty token", func(t *testing.T) {
		userService, _ := newTestUserService(t)

		_, err := userService.Authenticate(ctx, "")
		requireErrorIs(t, err, apperr.ErrUnauthorized)
	})

	t.Run("banned user", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)
		token := "session-token"
		user := &model.User{ID: 1, Username: testUsername, Status: model.UserStatusBanned}

		userRepository.EXPECT().
			GetUserBySessionTokenHash(ctx, security.HashSessionToken(token)).
			Return(user, nil)

		_, err := userService.Authenticate(ctx, token)
		requireErrorIs(t, err, apperr.ErrUserBanned)
	})
}

func TestLogout(t *testing.T) {
	ctx := context.Background()

	t.Run("revokes session", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)
		token := "session-token"

		userRepository.EXPECT().
			DeleteSessionByTokenHash(ctx, security.HashSessionToken(token)).
			Return(nil)

		if err := userService.Logout(ctx, token); err != nil {
			t.Fatalf("logout: %v", err)
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)
		token := "session-token"

		userRepository.EXPECT().
			DeleteSessionByTokenHash(ctx, security.HashSessionToken(token)).
			Return(nil).
			Times(2)

		if err := userService.Logout(ctx, token); err != nil {
			t.Fatalf("first logout: %v", err)
		}
		if err := userService.Logout(ctx, token); err != nil {
			t.Fatalf("second logout: %v", err)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		userService, _ := newTestUserService(t)

		requireErrorIs(t, userService.Logout(ctx, ""), apperr.ErrUnauthorized)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("valid credentials", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)
		user := &model.User{
			ID:           1,
			Username:     testUsername,
			PasswordHash: mustHashPassword(t, testPassword),
			Status:       model.UserStatusActive,
		}

		userRepository.EXPECT().
			GetUserByUsername(ctx, testUsername).
			Return(user, nil)

		var sessionTokenHash []byte
		userRepository.EXPECT().
			CreateSession(ctx, user, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *model.User, tokenHash []byte) error {
				sessionTokenHash = tokenHash
				return nil
			})

		token, err := userService.Login(ctx, service.LoginInput{
			Username: testUsername,
			Password: testPassword,
		})
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		if !bytes.Equal(sessionTokenHash, security.HashSessionToken(token)) {
			t.Fatal("stored session token hash does not match the issued token")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)
		user := &model.User{
			ID:           1,
			Username:     testUsername,
			PasswordHash: mustHashPassword(t, testPassword),
			Status:       model.UserStatusActive,
		}

		userRepository.EXPECT().
			GetUserByUsername(ctx, testUsername).
			Return(user, nil)

		_, err := userService.Login(ctx, service.LoginInput{
			Username: testUsername,
			Password: "WrongPassword123",
		})
		requireErrorIs(t, err, apperr.ErrInvalidCredentials)
	})

	t.Run("unknown user", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)

		userRepository.EXPECT().
			GetUserByUsername(ctx, testUsername).
			Return(nil, fmt.Errorf("user repository: %w", apperr.ErrUserNotFound))

		_, err := userService.Login(ctx, service.LoginInput{
			Username: testUsername,
			Password: testPassword,
		})
		requireErrorIs(t, err, apperr.ErrInvalidCredentials)
	})

	t.Run("banned user", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)
		user := &model.User{
			ID:           1,
			Username:     testUsername,
			PasswordHash: mustHashPassword(t, testPassword),
			Status:       model.UserStatusBanned,
		}

		userRepository.EXPECT().
			GetUserByUsername(ctx, testUsername).
			Return(user, nil)

		_, err := userService.Login(ctx, service.LoginInput{
			Username: testUsername,
			Password: testPassword,
		})
		requireErrorIs(t, err, apperr.ErrUserBanned)
	})
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("duplicate username", func(t *testing.T) {
		userService, userRepository := newTestUserService(t)

		userRepository.EXPECT().
			CreateUserWithSession(ctx, gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("user repository: %w", apperr.ErrUsernameTaken))

		_, err := userService.Register(ctx, service.RegisterUserInput{
			Username: testUsername,
			Password: testPassword,
		})
		requireErrorIs(t, err, apperr.ErrUsernameTaken)
	})
}
