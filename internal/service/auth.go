package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/queue"
	"github.com/PPRAMANIK62/devhunt/internal/queue/tasks"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo  repository.UserRepo
	jwtSecret string
	jwtExpiry time.Duration
	queue     *queue.Client // nil = no email sending
}

func NewAuthService(userRepo repository.UserRepo, secret string, expiryMinutes int, q *queue.Client) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: secret,
		jwtExpiry: time.Duration(expiryMinutes) * time.Minute,
		queue:     q,
	}
}

type RegisterInput struct {
	Email    string
	Password string
	Role     models.UserRole
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*models.User, error) {
	existing, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, apperr.Conflict("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, apperr.Internal("hash password", err)
	}

	user := &models.User{
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         input.Role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperr.Internal("create user", err)
	}

	// Generate and store verification token
	token, err := generateToken()
	if err != nil {
		return nil, apperr.Internal("generate verification token", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.userRepo.SetVerificationToken(ctx, user.ID, token, expiresAt); err != nil {
		return nil, err
	}

	// Enqueue verification email — don't fail registration if queue is unavailable
	if s.queue != nil {
		if err := s.queue.EnqueueEmailVerification(tasks.EmailVerificationPayload{
			Email: user.Email,
			Token: token,
		}); err != nil {
			slog.Error("failed to enqueue verification email", "error", err, "user_id", user.ID)
		}
	}

	return user, nil
}

type LoginOutput struct {
	Token string
	User  *models.User
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginOutput, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperr.Unauthorized("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, apperr.Unauthorized("invalid credentials")
	}

	if !user.EmailVerified {
		return nil, apperr.Forbidden("please verify your email before logging in")
	}

	token, err := s.createToken(user)
	if err != nil {
		return nil, apperr.Internal("create token", err)
	}

	return &LoginOutput{Token: token, User: user}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.userRepo.FindByVerificationToken(ctx, token)
	if err != nil {
		return apperr.NotFound("invalid or expired verification token")
	}

	if user.VerificationTokenExpiresAt != nil && time.Now().After(*user.VerificationTokenExpiresAt) {
		return apperr.Gone("verification token has expired")
	}

	return s.userRepo.SetVerified(ctx, user.ID)
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user.EmailVerified {
		// Silently succeed — don't leak whether the email exists or is verified
		return nil
	}

	token, err := generateToken()
	if err != nil {
		return apperr.Internal("generate verification token", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.userRepo.SetVerificationToken(ctx, user.ID, token, expiresAt); err != nil {
		return err
	}

	if s.queue != nil {
		if err := s.queue.EnqueueEmailVerification(tasks.EmailVerificationPayload{
			Email: user.Email,
			Token: token,
		}); err != nil {
			slog.Error("failed to enqueue verification email", "error", err, "user_id", user.ID)
		}
	}

	return nil
}

func (s *AuthService) createToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
