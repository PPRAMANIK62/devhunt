package models

import "time"

type UserRole string

const (
	RoleSeeker  UserRole = "seeker"
	RoleCompany UserRole = "company"
	RoleAdmin   UserRole = "admin"
)

type User struct {
	ID                          string     `json:"id"`
	Email                       string     `json:"email"`
	PasswordHash                string     `json:"-"`
	Role                        UserRole   `json:"role"`
	EmailVerified               bool       `json:"email_verified"`
	VerificationToken           *string    `json:"-"`
	VerificationTokenExpiresAt  *time.Time `json:"-"`
	CreatedAt                   time.Time  `json:"created_at"`
	UpdatedAt                   time.Time  `json:"updated_at"`
}
