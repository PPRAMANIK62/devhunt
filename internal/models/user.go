package models

import "time"

type UserRole string

const (
	RoleSeeker  UserRole = "seeker"
	RoleCompany UserRole = "company"
	RoleAdmin   UserRole = "admin"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" = never included in JSON output
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
