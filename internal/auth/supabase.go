package auth

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nedpals/supabase-go"
	"github.com/omkar273/codegeeky/internal/api/dto"
	"github.com/omkar273/codegeeky/internal/config"
	"github.com/omkar273/codegeeky/internal/domain/auth"
	ierr "github.com/omkar273/codegeeky/internal/errors"
	"github.com/omkar273/codegeeky/internal/logger"
	"github.com/omkar273/codegeeky/internal/security"
	"github.com/omkar273/codegeeky/internal/types"
)

type supabaseProvider struct {
	cfg               *config.Configuration
	supabase          *supabase.Client
	logger            *logger.Logger
	encryptionService security.EncryptionService
	jwks              *keyfunc.JWKS
}

func NewSupabaseProvider(cfg *config.Configuration, logger *logger.Logger, encryptionService security.EncryptionService) Provider {

	supabaseUrl := cfg.Supabase.URL
	secretKey := cfg.Supabase.SecretKey

	client := supabase.CreateClient(supabaseUrl, secretKey)

	if client == nil {
		log.Fatal("failed to create supabase client")
	}

	// Initialize JWKS client for JWT validation
	jwksUrl := cfg.Supabase.JWKSUrl
	if jwksUrl == "" {
		// Derive JWKS URL from Supabase URL if not provided
		jwksUrl = deriveJWKSUrl(supabaseUrl)
	}

	jwks, err := keyfunc.Get(jwksUrl, keyfunc.Options{
		RefreshTimeout:  time.Hour,
		RefreshInterval: time.Hour,
		RefreshErrorHandler: func(err error) {
			logger.Error("Failed to refresh JWKS", "error", err)
		},
	})
	if err != nil {
		log.Fatalf("failed to initialize JWKS client: %v", err)
	}

	return &supabaseProvider{
		cfg:               cfg,
		supabase:          client,
		logger:            logger,
		encryptionService: encryptionService,
		jwks:              jwks,
	}
}

func (p *supabaseProvider) GetProvider() types.AuthProvider {
	return types.AuthProviderSupabase
}

func (p *supabaseProvider) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
	// Parse and validate the JWT token using JWKS
	parsedToken, err := jwt.Parse(token, p.jwks.Keyfunc)

	if err != nil {
		p.logger.Error("Failed to parse JWT token", "error", err)
		return nil, ierr.NewErrorf("invalid token: %w", err).
			WithHint("Please use the correct token").
			Mark(ierr.ErrValidation)
	}

	// Check if token is valid
	if !parsedToken.Valid {
		p.logger.Error("JWT token is invalid")
		return nil, ierr.NewError("token is invalid").
			WithHint("Please use the correct token").
			Mark(ierr.ErrValidation)
	}

	// Extract claims from the token
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		p.logger.Error("Failed to extract claims from JWT token")
		return nil, ierr.NewError("failed to extract claims from token").
			WithHint("Please use the correct token").
			Mark(ierr.ErrValidation)
	}

	// Validate audience - should be "authenticated" for user tokens
	if aud, ok := claims["aud"].(string); ok && aud != "authenticated" {
		p.logger.Error("Invalid audience in JWT token", "audience", aud)
		return nil, ierr.NewError("invalid token audience").
			WithHint("Please use the correct token").
			Mark(ierr.ErrValidation)
	}

	// Validate role - should be "authenticated" for user tokens
	if role, ok := claims["role"].(string); ok && role != "authenticated" {
		p.logger.Error("Invalid role in JWT token", "role", role)
		return nil, ierr.NewError("invalid token role").
			WithHint("Please use the correct token").
			Mark(ierr.ErrValidation)
	}

	// Extract user information from JWT claims
	userID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	phone, _ := claims["phone"].(string)

	// Validate that we have at least a user ID
	if userID == "" {
		p.logger.Error("JWT token missing user ID (sub claim)")
		return nil, ierr.NewError("token missing user ID").
			WithHint("Please use the correct token").
			Mark(ierr.ErrValidation)
	}

	p.logger.Debug("Successfully validated JWT token", "user_id", userID, "email", email)

	return &auth.Claims{
		UserID: userID,
		Email:  email,
		Phone:  phone,
	}, nil
}

// SignUp is not used directly for Supabase as users sign up through the Supabase UI
// This method is kept for compatibility with the Provider interface
func (p *supabaseProvider) SignUp(ctx context.Context, req *dto.SignupRequest) (*dto.SignupResponse, error) {

	// For Supabase, we don't directly sign up users through this method
	// Instead, we validate the token and get user info
	// For Supabase, we validate the token and extract user info
	if req.AccessToken == "" {
		return nil, ierr.NewError("token is required").
			Mark(ierr.ErrPermissionDenied)
	}

	// Validate the token
	claims, err := p.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return nil, err
	}

	if claims.Email != req.Email {
		return nil, ierr.NewError("email mismatch").
			Mark(ierr.ErrPermissionDenied)
	}

	if claims.Phone != req.Phone {
		return nil, ierr.NewError("phone mismatch").
			Mark(ierr.ErrPermissionDenied)
	}

	// Create response
	resp := &dto.SignupResponse{
		ID:          claims.UserID,
		AccessToken: req.AccessToken,
	}

	return resp, nil
}

// GetUser gets user information by token
func (p *supabaseProvider) GetUser(ctx context.Context, userToken string) (*User, error) {
	supabaseUser, err := p.supabase.Auth.User(ctx, userToken)
	if err != nil {
		p.logger.Error("Failed to get user", "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert Supabase user to generic user
	user := &User{
		ID:        supabaseUser.ID,
		Email:     supabaseUser.Email,
		Phone:     "", // Phone is not available in Supabase User struct
		Metadata:  supabaseUser.UserMetadata,
		CreatedAt: supabaseUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: supabaseUser.UpdatedAt.Format(time.RFC3339),
	}

	return user, nil
}

// UpdateUser updates user information using the latest SDK method
func (p *supabaseProvider) UpdateUser(ctx context.Context, userToken string, updateData map[string]interface{}) (*User, error) {
	supabaseUser, err := p.supabase.Auth.UpdateUser(ctx, userToken, updateData)
	if err != nil {
		p.logger.Error("Failed to update user", "error", err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Convert Supabase user to generic user
	user := &User{
		ID:        supabaseUser.ID,
		Email:     supabaseUser.Email,
		Phone:     "", // Phone is not available in Supabase User struct
		Metadata:  supabaseUser.UserMetadata,
		CreatedAt: supabaseUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: supabaseUser.UpdatedAt.Format(time.RFC3339),
	}

	return user, nil
}

// deriveJWKSUrl derives the JWKS URL from a Supabase URL
func deriveJWKSUrl(supabaseUrl string) string {
	// Remove trailing slash if present
	url := strings.TrimSuffix(supabaseUrl, "/")
	// Append the JWKS endpoint (Supabase uses .well-known/jwks.json)
	return url + "/auth/v1/.well-known/jwks.json"
}

// Close cleans up resources used by the provider
func (p *supabaseProvider) Close() error {
	if p.jwks != nil {
		p.jwks.EndBackground()
	}
	return nil
}
