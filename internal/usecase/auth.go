package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/Ankesh2004/pontoon/internal/domain"
	gh "github.com/Ankesh2004/pontoon/internal/infrastructure/github"
)

type JWTClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type AuthUseCase struct {
	oauthService *gh.OAuthService
	userRepo     domain.UserRepository
	jwtSecret    string
}

func NewAuthUseCase(oauthService *gh.OAuthService, userRepo domain.UserRepository, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{
		oauthService: oauthService,
		userRepo:     userRepo,
		jwtSecret:    jwtSecret,
	}
}

func (uc *AuthUseCase) GetAuthURL(state string) string {
	return uc.oauthService.AuthURL(state)
}

func (uc *AuthUseCase) HandleCallback(ctx context.Context, code string) (string, error) {
	token, err := uc.oauthService.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("oauth exchange failed: %w", err)
	}

	client := gh.NewClient(ctx, token)

	ghUser, err := client.GetUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get github user: %w", err)
	}

	user := &domain.User{
		ID:          uuid.New().String(),
		GitHubID:    ghUser.GitHubID,
		GitHubUsername: ghUser.GitHubUsername,
		Email:       ghUser.Email,
		AvatarURL:   ghUser.AvatarURL,
		AccessToken: token.AccessToken,
	}

	if err := uc.userRepo.UpsertByGitHubID(user); err != nil {
		return "", fmt.Errorf("failed to upsert user: %w", err)
	}

	jwtToken, err := uc.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return jwtToken, nil
}

func (uc *AuthUseCase) generateJWT(user *domain.User) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		TenantID: user.ID,
		Username: user.GitHubUsername,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "pontoon",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *AuthUseCase) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(uc.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
