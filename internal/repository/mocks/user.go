package mocks

import (
	"context"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/models"
)

type UserRepo struct {
	CreateFn                    func(ctx context.Context, user *models.User) error
	FindByEmailFn               func(ctx context.Context, email string) (*models.User, error)
	FindByIDFn                  func(ctx context.Context, id string) (*models.User, error)
	FindByVerificationTokenFn   func(ctx context.Context, token string) (*models.User, error)
	SetVerificationTokenFn      func(ctx context.Context, userID, token string, expiresAt time.Time) error
	SetVerifiedFn               func(ctx context.Context, userID string) error
}

func (m *UserRepo) Create(ctx context.Context, user *models.User) error {
	if m.CreateFn == nil {
		panic("mocks.UserRepo.CreateFn not set")
	}
	return m.CreateFn(ctx, user)
}

func (m *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.FindByEmailFn == nil {
		panic("mocks.UserRepo.FindByEmailFn not set")
	}
	return m.FindByEmailFn(ctx, email)
}

func (m *UserRepo) FindByID(ctx context.Context, id string) (*models.User, error) {
	if m.FindByIDFn == nil {
		panic("mocks.UserRepo.FindByIDFn not set")
	}
	return m.FindByIDFn(ctx, id)
}

func (m *UserRepo) FindByVerificationToken(ctx context.Context, token string) (*models.User, error) {
	if m.FindByVerificationTokenFn == nil {
		panic("mocks.UserRepo.FindByVerificationTokenFn not set")
	}
	return m.FindByVerificationTokenFn(ctx, token)
}

func (m *UserRepo) SetVerificationToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	if m.SetVerificationTokenFn == nil {
		panic("mocks.UserRepo.SetVerificationTokenFn not set")
	}
	return m.SetVerificationTokenFn(ctx, userID, token, expiresAt)
}

func (m *UserRepo) SetVerified(ctx context.Context, userID string) error {
	if m.SetVerifiedFn == nil {
		panic("mocks.UserRepo.SetVerifiedFn not set")
	}
	return m.SetVerifiedFn(ctx, userID)
}
