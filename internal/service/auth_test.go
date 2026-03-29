package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository/mocks"
	"github.com/PPRAMANIK62/devhunt/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newAuthSvc(userRepo *mocks.UserRepo) *service.AuthService {
	return service.NewAuthService(userRepo, "test-secret", 60, nil)
}

func TestAuthService_Register_EmailConflict(t *testing.T) {
	userRepo := &mocks.UserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			// Email already exists
			return &models.User{Email: email}, nil
		},
	}
	svc := newAuthSvc(userRepo)

	_, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "taken@example.com",
		Password: "password123",
		Role:     models.RoleSeeker,
	})

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeConflict, appErr.Type)
}

func TestAuthService_Register_Success(t *testing.T) {
	var createdUser *models.User
	var storedToken string

	userRepo := &mocks.UserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return nil, apperr.NotFound("user")
		},
		CreateFn: func(ctx context.Context, user *models.User) error {
			user.ID = "user-new"
			createdUser = user
			return nil
		},
		SetVerificationTokenFn: func(ctx context.Context, userID, token string, expiresAt time.Time) error {
			storedToken = token
			return nil
		},
	}
	svc := newAuthSvc(userRepo)

	user, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "new@example.com",
		Password: "password123",
		Role:     models.RoleSeeker,
	})

	require.NoError(t, err)
	assert.Equal(t, "user-new", user.ID)
	assert.Equal(t, "new@example.com", createdUser.Email)
	assert.NotEmpty(t, storedToken)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := &mocks.UserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return nil, apperr.NotFound("user")
		},
	}
	svc := newAuthSvc(userRepo)

	_, err := svc.Login(context.Background(), "nobody@example.com", "password123")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeUnauthorized, appErr.Type)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)

	userRepo := &mocks.UserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{ID: "u1", Email: email, PasswordHash: string(hash), EmailVerified: true}, nil
		},
	}
	svc := newAuthSvc(userRepo)

	_, err := svc.Login(context.Background(), "user@example.com", "wrong-password")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeUnauthorized, appErr.Type)
}

func TestAuthService_Login_EmailNotVerified(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)

	userRepo := &mocks.UserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{ID: "u1", Email: email, PasswordHash: string(hash), EmailVerified: false}, nil
		},
	}
	svc := newAuthSvc(userRepo)

	_, err := svc.Login(context.Background(), "user@example.com", "password123")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeForbidden, appErr.Type)
}

func TestAuthService_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)

	userRepo := &mocks.UserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{
				ID:            "u1",
				Email:         email,
				PasswordHash:  string(hash),
				Role:          models.RoleSeeker,
				EmailVerified: true,
			}, nil
		},
	}
	svc := newAuthSvc(userRepo)

	out, err := svc.Login(context.Background(), "user@example.com", "password123")

	require.NoError(t, err)
	assert.NotEmpty(t, out.Token)
	assert.Equal(t, "u1", out.User.ID)
}

func TestAuthService_VerifyEmail_InvalidToken(t *testing.T) {
	userRepo := &mocks.UserRepo{
		FindByVerificationTokenFn: func(ctx context.Context, token string) (*models.User, error) {
			return nil, apperr.NotFound("verification token")
		},
	}
	svc := newAuthSvc(userRepo)

	err := svc.VerifyEmail(context.Background(), "bad-token")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}

func TestAuthService_VerifyEmail_ExpiredToken(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)

	userRepo := &mocks.UserRepo{
		FindByVerificationTokenFn: func(ctx context.Context, token string) (*models.User, error) {
			return &models.User{ID: "u1", VerificationTokenExpiresAt: &past}, nil
		},
	}
	svc := newAuthSvc(userRepo)

	err := svc.VerifyEmail(context.Background(), "expired-token")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeGone, appErr.Type)
}

func TestAuthService_VerifyEmail_Success(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	verified := false

	userRepo := &mocks.UserRepo{
		FindByVerificationTokenFn: func(ctx context.Context, token string) (*models.User, error) {
			return &models.User{ID: "u1", VerificationTokenExpiresAt: &future}, nil
		},
		SetVerifiedFn: func(ctx context.Context, userID string) error {
			verified = true
			return nil
		},
	}
	svc := newAuthSvc(userRepo)

	err := svc.VerifyEmail(context.Background(), "valid-token")

	require.NoError(t, err)
	assert.True(t, verified)
}
