package repository_test

import (
	"context"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/repository"
)

func TestUserRepositoryCreateUser(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(testPool)

	t.Run("creates an active user", func(t *testing.T) {
		resetDatabase(t)

		user := &model.User{
			Username:     testUserUsername,
			PasswordHash: testUserPassword,
			Email:        testUserEmail,
		}
		if err := userRepository.CreateUserWithSession(ctx, user, testTokenHash(user.Username)); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if user.ID == 0 {
			t.Fatal("user ID was not assigned")
		}

		got, err := userRepository.GetUserByUsername(ctx, testUserUsername)
		if err != nil {
			t.Fatalf("get user: %v", err)
		}
		if got.ID != user.ID || got.Username != user.Username {
			t.Fatalf("user = %+v, want %+v", got, user)
		}
		if got.PasswordHash != testUserPassword {
			t.Fatalf("password hash = %q, want %q", got.PasswordHash, testUserPassword)
		}
		if got.Email != testUserEmail {
			t.Fatalf("email = %q, want %q", got.Email, testUserEmail)
		}
		if got.Status != model.UserStatusActive {
			t.Fatalf("status = %q, want %q", got.Status, model.UserStatusActive)
		}
		if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
			t.Fatalf("timestamps are not set: %+v", got)
		}
	})

	t.Run("stores a missing email as an empty string", func(t *testing.T) {
		resetDatabase(t)

		user := &model.User{
			Username:     testUserUsername,
			PasswordHash: testUserPassword,
		}
		if err := userRepository.CreateUserWithSession(ctx, user, testTokenHash(user.Username)); err != nil {
			t.Fatalf("create user: %v", err)
		}

		got, err := userRepository.GetUserByUsername(ctx, testUserUsername)
		if err != nil {
			t.Fatalf("get user: %v", err)
		}
		if got.Email != "" {
			t.Fatalf("email = %q, want empty", got.Email)
		}
	})

	t.Run("rejects a duplicate username", func(t *testing.T) {
		resetDatabase(t)
		insertUser(t)

		err := userRepository.CreateUserWithSession(ctx, &model.User{
			Username:     testUserUsername,
			PasswordHash: testUserPassword,
			Email:        "another@example.com",
		}, testTokenHash("duplicate-username"))
		requireErrorIs(t, err, apperr.ErrUsernameTaken)
	})

	t.Run("rejects a duplicate email", func(t *testing.T) {
		resetDatabase(t)
		insertUser(t)

		err := userRepository.CreateUserWithSession(ctx, &model.User{
			Username:     "alice02",
			PasswordHash: testUserPassword,
			Email:        testUserEmail,
		}, testTokenHash("duplicate-email"))
		requireErrorIs(t, err, apperr.ErrEmailTaken)
	})

	t.Run("returns not found for an unknown username", func(t *testing.T) {
		resetDatabase(t)

		_, err := userRepository.GetUserByUsername(ctx, testUserUsername)
		requireErrorIs(t, err, apperr.ErrUserNotFound)
	})
}

func TestUserRepositoryCreateUserWithSession(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(testPool)

	t.Run("creates the user and its session", func(t *testing.T) {
		resetDatabase(t)

		user := &model.User{
			Username:     testUserUsername,
			PasswordHash: testUserPassword,
			Email:        testUserEmail,
		}
		tokenHash := []byte("user-with-session-token")
		if err := userRepository.CreateUserWithSession(ctx, user, tokenHash); err != nil {
			t.Fatalf("create user with session: %v", err)
		}

		got, err := userRepository.GetUserBySessionTokenHash(ctx, tokenHash)
		if err != nil {
			t.Fatalf("get user by session: %v", err)
		}
		if got.ID != user.ID {
			t.Fatalf("user ID = %d, want %d", got.ID, user.ID)
		}
	})

	t.Run("rolls back the user when the session cannot be created", func(t *testing.T) {
		resetDatabase(t)
		otherUser := insertUser(t)
		tokenHash := []byte("duplicate-session-token")
		if err := userRepository.CreateSession(ctx, otherUser, tokenHash); err != nil {
			t.Fatalf("create session: %v", err)
		}

		user := &model.User{
			Username:     "alice02",
			PasswordHash: testUserPassword,
			Email:        "alice02@example.com",
		}
		if err := userRepository.CreateUserWithSession(ctx, user, tokenHash); err == nil {
			t.Fatal("create user with session: expected an error")
		}

		_, err := userRepository.GetUserByUsername(ctx, "alice02")
		requireErrorIs(t, err, apperr.ErrUserNotFound)
	})
}

func TestUserRepositorySessions(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(testPool)

	t.Run("creates and resolves a session", func(t *testing.T) {
		resetDatabase(t)
		user := insertUser(t)
		tokenHash := []byte("user-session-token")

		if err := userRepository.CreateSession(ctx, user, tokenHash); err != nil {
			t.Fatalf("create session: %v", err)
		}

		got, err := userRepository.GetUserBySessionTokenHash(ctx, tokenHash)
		if err != nil {
			t.Fatalf("get user by session: %v", err)
		}
		if got.ID != user.ID {
			t.Fatalf("user ID = %d, want %d", got.ID, user.ID)
		}
	})

	t.Run("rejects an unknown session", func(t *testing.T) {
		resetDatabase(t)

		_, err := userRepository.GetUserBySessionTokenHash(ctx, []byte("unknown-token"))
		requireErrorIs(t, err, apperr.ErrInvalidSession)
	})

	t.Run("rejects an expired session", func(t *testing.T) {
		resetDatabase(t)
		user := insertUser(t)
		tokenHash := []byte("expired-session-token")

		if err := userRepository.CreateSession(ctx, user, tokenHash); err != nil {
			t.Fatalf("create session: %v", err)
		}
		if _, err := testPool.Exec(ctx, `UPDATE sessions SET expires_at = now() - INTERVAL '1 hour'`); err != nil {
			t.Fatalf("expire session: %v", err)
		}

		_, err := userRepository.GetUserBySessionTokenHash(ctx, tokenHash)
		requireErrorIs(t, err, apperr.ErrInvalidSession)
	})

	t.Run("deletes a session", func(t *testing.T) {
		resetDatabase(t)
		user := insertUser(t)
		tokenHash := []byte("deleted-session-token")

		if err := userRepository.CreateSession(ctx, user, tokenHash); err != nil {
			t.Fatalf("create session: %v", err)
		}
		if err := userRepository.DeleteSessionByTokenHash(ctx, tokenHash); err != nil {
			t.Fatalf("delete session: %v", err)
		}

		_, err := userRepository.GetUserBySessionTokenHash(ctx, tokenHash)
		requireErrorIs(t, err, apperr.ErrInvalidSession)
	})
}
