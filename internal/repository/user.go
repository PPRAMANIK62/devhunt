package repository

import (
	"context"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `id, email, password_hash, role, email_verified, verification_token, verification_token_expires_at, created_at, updated_at`

func scanUser(row pgx.Row, user *models.User) error {
	return row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.EmailVerified,
		&user.VerificationToken,
		&user.VerificationTokenExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, user.Email, user.PasswordHash, user.Role).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email), user)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("user")
	}
	if err != nil {
		return nil, apperr.Internal("find user", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id), user)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("user")
	}
	if err != nil {
		return nil, apperr.Internal("find user", err)
	}
	return user, nil
}

func (r *UserRepository) FindByVerificationToken(ctx context.Context, token string) (*models.User, error) {
	user := &models.User{}
	err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE verification_token = $1`, token), user)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("verification token")
	}
	if err != nil {
		return nil, apperr.Internal("find user by token", err)
	}
	return user, nil
}

func (r *UserRepository) SetVerificationToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET verification_token = $1, verification_token_expires_at = $2, updated_at = NOW()
		WHERE id = $3
	`, token, expiresAt, userID)
	if err != nil {
		return apperr.Internal("set verification token", err)
	}
	return nil
}

func (r *UserRepository) SetVerified(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET email_verified = true, verification_token = NULL, verification_token_expires_at = NULL, updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return apperr.Internal("set user verified", err)
	}
	return nil
}
