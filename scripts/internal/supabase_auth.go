package internal

import (
	"context"

	"github.com/nedpals/supabase-go"
	"github.com/omkar273/nashikdarshan/internal/config"
	"github.com/omkar273/nashikdarshan/internal/domain/auth"
	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/omkar273/nashikdarshan/internal/logger"
)

// go run scripts/main.go -cmd onboard-user --email user@example.com --password pass123

// SupabaseAuthHelper provides dev-only authentication methods for scripts
type SupabaseAuthHelper struct {
	supabase *supabase.Client
	logger   *logger.Logger
}

// NewSupabaseAuthHelper creates a new helper for dev-only auth operations
func NewSupabaseAuthHelper(cfg *config.Configuration, logger *logger.Logger) *SupabaseAuthHelper {
	client := supabase.CreateClient(cfg.Supabase.URL, cfg.Supabase.SecretKey)
	if client == nil {
		logger.Fatal("failed to create supabase client")
		return nil
	}

	return &SupabaseAuthHelper{
		supabase: client,
		logger:   logger,
	}
}

// SignInWithPassword authenticates a user with email and password
// Returns access token and user claims if successful
func (h *SupabaseAuthHelper) SignInWithPassword(ctx context.Context, email, password string) (*auth.Claims, string, error) {
	if email == "" {
		return nil, "", ierr.NewError("email is required").
			WithHint("Please provide a valid email address").
			Mark(ierr.ErrValidation)
	}
	if password == "" {
		return nil, "", ierr.NewError("password is required").
			WithHint("Please provide a password").
			Mark(ierr.ErrValidation)
	}

	// Sign in with email and password
	authResponse, err := h.supabase.Auth.SignIn(ctx, supabase.UserCredentials{
		Email:    email,
		Password: password,
	})
	if err != nil {
		h.logger.Error("Failed to sign in", "email", email, "error", err)
		return nil, "", ierr.NewErrorf("failed to sign in: %w", err).
			WithHint("Invalid email or password").
			Mark(ierr.ErrPermissionDenied)
	}

	claims := &auth.Claims{
		UserID: authResponse.User.ID,
		Email:  authResponse.User.Email,
	}

	accessToken := authResponse.AccessToken
	h.logger.Debug("User signed in", "user_id", claims.UserID, "email", claims.Email)
	return claims, accessToken, nil
}

// SignUpWithPassword creates a new user with email and password
// Returns access token and user claims if successful
func (h *SupabaseAuthHelper) SignUpWithPassword(ctx context.Context, email, password, name string) (*auth.Claims, string, error) {
	if email == "" {
		return nil, "", ierr.NewError("email is required").
			WithHint("Please provide a valid email address").
			Mark(ierr.ErrValidation)
	}
	if password == "" {
		return nil, "", ierr.NewError("password is required").
			WithHint("Please provide a password").
			Mark(ierr.ErrValidation)
	}

	// Prepare signup data
	signupData := supabase.UserCredentials{
		Email:    email,
		Password: password,
	}

	// Sign up with email and password
	// Note: SignUp returns *User, not AuthenticatedDetails, so we need to sign in after to get token
	supabaseUser, err := h.supabase.Auth.SignUp(ctx, signupData)
	if err != nil {
		h.logger.Error("Failed to sign up", "email", email, "error", err)
		return nil, "", ierr.NewErrorf("failed to sign up: %w", err).
			WithHint("Failed to create user account").
			Mark(ierr.ErrInternal)
	}

	// After signup, sign in to get access token
	authResponse, err := h.supabase.Auth.SignIn(ctx, signupData)
	if err != nil {
		h.logger.Error("Failed to sign in after signup", "email", email, "error", err)
		return nil, "", ierr.NewErrorf("failed to sign in after signup: %w", err).
			WithHint("User created but failed to get access token").
			Mark(ierr.ErrInternal)
	}

	// Update user metadata with name if provided
	if name != "" {
		_, err = h.supabase.Auth.UpdateUser(ctx, authResponse.AccessToken, map[string]interface{}{
			"user_metadata": map[string]interface{}{
				"name": name,
			},
		})
		if err != nil {
			h.logger.Warn("Failed to update user metadata with name", "error", err)
			// Don't fail the whole operation if metadata update fails
		}
	}

	claims := &auth.Claims{
		UserID: supabaseUser.ID,
		Email:  supabaseUser.Email,
	}

	accessToken := authResponse.AccessToken
	h.logger.Debug("User signed up", "user_id", claims.UserID, "email", claims.Email)
	return claims, accessToken, nil
}
