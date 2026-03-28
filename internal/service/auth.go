package service

import (
	"context"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
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
	userRepo  *repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, secret string, expiryMinutes int) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: secret,
		jwtExpiry: time.Duration(expiryMinutes) * time.Minute,
	}
}

type RegisterInput struct {
	Email    string
	Password string
	Role     models.UserRole
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*models.User, error) {
	// Check duplicate email
	existing, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, apperr.Conflict("email already registered")
	}

	// bcrypt cost 12 - slow enough to make brute force impractical
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

	return user, nil
}

type LoginOutput struct {
	Token string
	User  *models.User
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginOutput, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Generic message - never say "email not found" (user enumeration)
		return nil, apperr.Unauthorized("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, apperr.Unauthorized("invalid credentials")
	}

	token, err := s.createToken(user)
	if err != nil {
		return nil, apperr.Internal("create token", err)
	}

	return &LoginOutput{
		Token: token,
		User:  user,
	}, nil
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

	return jwt.
		NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(s.jwtSecret))
}
