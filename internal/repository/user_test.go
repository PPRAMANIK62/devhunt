package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndFindByEmail(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepository(testDB)

	email := fmt.Sprintf("user+%d@test.com", time.Now().UnixNano())
	user := &models.User{
		Email:        email,
		PasswordHash: "hashed-password",
		Role:         models.RoleSeeker,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.NotZero(t, user.CreatedAt)

	found, err := repo.FindByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, email, found.Email)
	assert.Equal(t, models.RoleSeeker, found.Role)
	assert.False(t, found.EmailVerified)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	repo := repository.NewUserRepository(testDB)
	_, err := repo.FindByEmail(context.Background(), "nobody@nowhere.com")
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	repo := repository.NewUserRepository(testDB)
	_, err := repo.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepository(testDB)

	email := fmt.Sprintf("dup+%d@test.com", time.Now().UnixNano())

	err := repo.Create(ctx, &models.User{Email: email, PasswordHash: "hash", Role: models.RoleSeeker})
	require.NoError(t, err)

	// Second create with same email should fail (unique constraint)
	err = repo.Create(ctx, &models.User{Email: email, PasswordHash: "hash2", Role: models.RoleSeeker})
	require.Error(t, err)
}

func TestUserRepository_SetVerificationToken(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepository(testDB)

	email := fmt.Sprintf("verify+%d@test.com", time.Now().UnixNano())
	user := &models.User{Email: email, PasswordHash: "hash", Role: models.RoleSeeker}
	require.NoError(t, repo.Create(ctx, user))

	token := "test-token-abc"
	expiresAt := time.Now().Add(24 * time.Hour)
	err := repo.SetVerificationToken(ctx, user.ID, token, expiresAt)
	require.NoError(t, err)

	found, err := repo.FindByVerificationToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.NotNil(t, found.VerificationToken)
	assert.Equal(t, token, *found.VerificationToken)
}

func TestUserRepository_SetVerified(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepository(testDB)

	email := fmt.Sprintf("setverified+%d@test.com", time.Now().UnixNano())
	user := &models.User{Email: email, PasswordHash: "hash", Role: models.RoleSeeker}
	require.NoError(t, repo.Create(ctx, user))

	// Set a token first
	require.NoError(t, repo.SetVerificationToken(ctx, user.ID, "some-token", time.Now().Add(time.Hour)))

	err := repo.SetVerified(ctx, user.ID)
	require.NoError(t, err)

	found, err := repo.FindByEmail(ctx, email)
	require.NoError(t, err)
	assert.True(t, found.EmailVerified)
	assert.Nil(t, found.VerificationToken)
	assert.Nil(t, found.VerificationTokenExpiresAt)
}

func TestUserRepository_FindByVerificationToken_NotFound(t *testing.T) {
	repo := repository.NewUserRepository(testDB)
	_, err := repo.FindByVerificationToken(context.Background(), "nonexistent-token")
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}
