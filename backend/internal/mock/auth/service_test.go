package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testRepository struct {
	account     *Account
	sessions    map[string]time.Time
	deleted     int
	lastCreated time.Time
}

func (r *testRepository) UpsertAccount(
	_ context.Context,
	username string,
	passwordHash string,
) (int64, error) {
	r.account = &Account{
		ID:           1,
		Username:     username,
		PasswordHash: passwordHash,
	}
	return r.account.ID, nil
}

func (r *testRepository) GetAccountByUsername(
	_ context.Context,
	username string,
) (*Account, error) {
	if r.account == nil || r.account.Username != username {
		return nil, errAccountNotFound
	}
	return r.account, nil
}

func (r *testRepository) CreateSession(
	_ context.Context,
	_ int64,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	r.sessions[string(tokenHash)] = expiresAt
	r.lastCreated = expiresAt
	return nil
}

func (r *testRepository) GetAccountBySessionTokenHash(
	_ context.Context,
	tokenHash []byte,
) (*Account, error) {
	expiresAt, ok := r.sessions[string(tokenHash)]
	if !ok || !expiresAt.After(time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)) {
		return nil, ErrInvalidSession
	}
	return r.account, nil
}

func (r *testRepository) DeleteSessionByTokenHash(
	_ context.Context,
	tokenHash []byte,
) error {
	delete(r.sessions, string(tokenHash))
	r.deleted++
	return nil
}

func (r *testRepository) DeleteSessionsByAccountID(context.Context, int64) error {
	r.sessions = make(map[string]time.Time)
	return nil
}

func (r *testRepository) DeleteOtherAccounts(context.Context, string) error {
	return nil
}

func (r *testRepository) DeleteExpiredSessions(context.Context) error {
	return nil
}

func TestServiceSessionLifecycle(t *testing.T) {
	repository := &testRepository{sessions: make(map[string]time.Time)}
	service, err := NewServiceWithRepository(repository, Config{
		Username:   "mock-operator",
		Password:   "mockpay-test-password",
		SessionTTL: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	}

	if err := service.Bootstrap(context.Background()); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	token, err := service.Login(
		context.Background(),
		"mock-operator",
		"mockpay-test-password",
	)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" || !repository.lastCreated.Equal(service.now().Add(2*time.Hour)) {
		t.Fatalf("token = %q, expires at = %s", token, repository.lastCreated)
	}

	account, err := service.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if account.Username != "mock-operator" {
		t.Fatalf("username = %q", account.Username)
	}

	if err := service.Logout(context.Background(), token); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := service.Authenticate(context.Background(), token); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("authenticate after logout error = %v", err)
	}
}

func TestServiceRejectsInvalidCredentials(t *testing.T) {
	repository := &testRepository{sessions: make(map[string]time.Time)}
	service, err := NewServiceWithRepository(repository, Config{
		Username: "mock-operator",
		Password: "mockpay-test-password",
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.Bootstrap(context.Background()); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	for _, input := range []struct {
		username string
		password string
	}{
		{username: "unknown", password: "mockpay-test-password"},
		{username: "mock-operator", password: "mocklogistics-test-password"},
	} {
		if _, err := service.Login(
			context.Background(),
			input.username,
			input.password,
		); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("login(%q) error = %v", input.username, err)
		}
	}
}

func TestBootstrapRotatesCredentialsAndRevokesSessions(t *testing.T) {
	repository := &testRepository{sessions: make(map[string]time.Time)}
	initial, err := NewServiceWithRepository(repository, Config{
		Username: "mock-operator",
		Password: "mockpay-test-password",
	})
	if err != nil {
		t.Fatalf("new initial service: %v", err)
	}
	if err := initial.Bootstrap(context.Background()); err != nil {
		t.Fatalf("bootstrap initial service: %v", err)
	}
	token, err := initial.Login(
		context.Background(),
		"mock-operator",
		"mockpay-test-password",
	)
	if err != nil {
		t.Fatalf("initial login: %v", err)
	}

	rotated, err := NewServiceWithRepository(repository, Config{
		Username: "mock-operator",
		Password: "mocklogistics-test-password",
	})
	if err != nil {
		t.Fatalf("new rotated service: %v", err)
	}
	if err := rotated.Bootstrap(context.Background()); err != nil {
		t.Fatalf("bootstrap rotated service: %v", err)
	}
	if _, err := rotated.Authenticate(
		context.Background(),
		token,
	); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("old session error = %v", err)
	}
	if _, err := rotated.Login(
		context.Background(),
		"mock-operator",
		"mockpay-test-password",
	); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("old password error = %v", err)
	}
	if _, err := rotated.Login(
		context.Background(),
		"mock-operator",
		"mocklogistics-test-password",
	); err != nil {
		t.Fatalf("rotated login: %v", err)
	}
}
