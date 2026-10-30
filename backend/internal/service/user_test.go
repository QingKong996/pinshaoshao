package service_test

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/service"
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

func (r *fakeUserRepository) SessionCount() int {
	return len(r.sessions)
}

func newTestUserService() (*service.UserService, *fakeUserRepository) {
	userRepository := newFakeUserRepository()
	return service.NewUserService(userRepository), userRepository
}

func register(t *testing.T, userService *service.UserService) string {
	t.Helper()

	token, err := userService.Register(context.Background(), &dto.RegisterRequest{
		Username: testUsername,
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return token
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
		userService, _ := newTestUserService()
		token := register(t, userService)

		user, err := userService.Authenticate(ctx, token)
		if err != nil {
			t.Fatalf("authenticate: %v", err)
		}
		if user.Username != testUsername {
			t.Fatalf("username = %q, want %q", user.Username, testUsername)
		}
	})

	t.Run("unknown token", func(t *testing.T) {
		userService, _ := newTestUserService()

		_, err := userService.Authenticate(ctx, "not-a-real-token")
		requireErrorIs(t, err, apperr.ErrInvalidSession)
	})

	t.Run("empty token", func(t *testing.T) {
		userService, _ := newTestUserService()

		_, err := userService.Authenticate(ctx, "")
		requireErrorIs(t, err, apperr.ErrUnauthorized)
	})

	t.Run("banned user", func(t *testing.T) {
		userService, userRepository := newTestUserService()
		token := register(t, userService)

		userRepository.SetUserStatus(testUsername, model.UserStatusBanned)

		_, err := userService.Authenticate(ctx, token)
		requireErrorIs(t, err, apperr.ErrUserBanned)
	})
}

func TestLogout(t *testing.T) {
	ctx := context.Background()

	t.Run("revokes session", func(t *testing.T) {
		userService, userRepository := newTestUserService()
		token := register(t, userService)

		if err := userService.Logout(ctx, token); err != nil {
			t.Fatalf("logout: %v", err)
		}
		if got := userRepository.SessionCount(); got != 0 {
			t.Fatalf("session count = %d, want 0", got)
		}

		_, err := userService.Authenticate(ctx, token)
		requireErrorIs(t, err, apperr.ErrInvalidSession)
	})

	t.Run("is idempotent", func(t *testing.T) {
		userService, _ := newTestUserService()
		token := register(t, userService)

		if err := userService.Logout(ctx, token); err != nil {
			t.Fatalf("first logout: %v", err)
		}
		if err := userService.Logout(ctx, token); err != nil {
			t.Fatalf("second logout: %v", err)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		userService, _ := newTestUserService()

		requireErrorIs(t, userService.Logout(ctx, ""), apperr.ErrUnauthorized)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("valid credentials", func(t *testing.T) {
		userService, _ := newTestUserService()
		register(t, userService)

		token, err := userService.Login(ctx, &dto.LoginRequest{
			Username: testUsername,
			Password: testPassword,
		})
		if err != nil {
			t.Fatalf("login: %v", err)
		}

		user, err := userService.Authenticate(ctx, token)
		if err != nil {
			t.Fatalf("authenticate: %v", err)
		}
		if user.Username != testUsername {
			t.Fatalf("username = %q, want %q", user.Username, testUsername)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		userService, _ := newTestUserService()
		register(t, userService)

		_, err := userService.Login(ctx, &dto.LoginRequest{
			Username: testUsername,
			Password: "WrongPassword123",
		})
		requireErrorIs(t, err, apperr.ErrInvalidCredentials)
	})

	t.Run("unknown user", func(t *testing.T) {
		userService, _ := newTestUserService()

		_, err := userService.Login(ctx, &dto.LoginRequest{
			Username: testUsername,
			Password: testPassword,
		})
		requireErrorIs(t, err, apperr.ErrInvalidCredentials)
	})

	t.Run("banned user", func(t *testing.T) {
		userService, userRepository := newTestUserService()
		register(t, userService)

		userRepository.SetUserStatus(testUsername, model.UserStatusBanned)

		_, err := userService.Login(ctx, &dto.LoginRequest{
			Username: testUsername,
			Password: testPassword,
		})
		requireErrorIs(t, err, apperr.ErrUserBanned)
	})
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("duplicate username", func(t *testing.T) {
		userService, _ := newTestUserService()
		register(t, userService)

		_, err := userService.Register(ctx, &dto.RegisterRequest{
			Username: testUsername,
			Password: testPassword,
		})
		requireErrorIs(t, err, apperr.ErrUsernameTaken)
	})
}
