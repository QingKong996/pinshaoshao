package model

import "time"

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
	UserStatusBanned UserStatus = "banned"
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Email        string
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
