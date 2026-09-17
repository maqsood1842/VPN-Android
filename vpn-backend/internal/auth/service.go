package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vpn/backend/internal/user"
	"github.com/vpn/backend/pkg/config"
	"github.com/vpn/backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type Service struct {
	userRepo *user.Repository
	jwtCfg   *config.JWTConfig
}

func NewService(userRepo *user.Repository, jwtCfg *config.JWTConfig) *Service {
	return &Service{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

type SignupInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User  *user.User   `json:"user"`
	Tokens *jwt.TokenPair `json:"tokens"`
}

func (s *Service) Signup(ctx context.Context, input SignupInput) (*AuthResponse, error) {
	existing, _ := s.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	tokens, err := jwt.GenerateTokens(u.ID.String(), u.Email, s.jwtCfg)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{User: u, Tokens: tokens}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := jwt.GenerateTokens(u.ID.String(), u.Email, s.jwtCfg)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{User: u, Tokens: tokens}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	claims, err := jwt.ValidateToken(refreshToken, s.jwtCfg)
	if err != nil {
		return nil, err
	}

	return jwt.GenerateTokens(claims.UserID, claims.Email, s.jwtCfg)
}