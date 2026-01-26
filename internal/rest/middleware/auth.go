package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omkar273/nashikdarshan/internal/auth"
	"github.com/omkar273/nashikdarshan/internal/config"
	"github.com/omkar273/nashikdarshan/internal/logger"
	"github.com/omkar273/nashikdarshan/internal/service"
	"github.com/omkar273/nashikdarshan/internal/types"
)

// validateAPIKey validates the API key and returns userID and permissions if valid
// Validates against the database via SecretService
func validateAPIKey(secretService service.SecretService, apiKey string) (userID string, permissions []string, valid bool) {
	if apiKey == "" {
		return "", nil, false
	}

	// Validate against database via SecretService
	if secretService != nil {
		userID, permissions, valid = secretService.ValidateAPIKey(context.Background(), apiKey)
		if valid {
			return userID, permissions, true
		}
	}

	return "", nil, false
}

// setContextValues sets the user ID, user email, and permissions in the context
func setContextValues(c *gin.Context, userID, userEmail string, permissions []string) {
	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, types.CtxUserID, userID)
	if userEmail != "" {
		ctx = context.WithValue(ctx, types.CtxUserEmail, userEmail)
	}
	// Store permissions for future RBAC implementation
	if permissions != nil {
		ctx = context.WithValue(ctx, types.CtxPermissions, permissions)
	}

	c.Request = c.Request.WithContext(ctx)
}

// GuestAuthenticateMiddleware is a middleware that allows requests without authentication
// For now it sets a default user ID and user email in the request context
func GuestAuthenticateMiddleware(c *gin.Context) {
	c.Next()
}

// AuthenticateMiddleware is a middleware that authenticates requests based on either:
// 1. API key in the X-API-Key header
// 2. JWT token in the Authorization header as a Bearer token
func AuthenticateMiddleware(cfg *config.Configuration, logger *logger.Logger, secretService service.SecretService) gin.HandlerFunc {
	authProvider := auth.NewSupabaseProvider(cfg, logger)

	return func(c *gin.Context) {
		// First check for API key
		apiKey := c.GetHeader(types.HeaderAPIKey)
		userID, permissions, valid := validateAPIKey(secretService, apiKey)
		if valid {
			// API key authentication - no email available, set empty string
			setContextValues(c, userID, "", permissions)
			c.Next()
			return
		}

		// If no API key or invalid, check for JWT token
		authHeader := c.GetHeader(types.HeaderAuthorization)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Check if the authorization header is in the correct format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := authProvider.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			logger.Errorw("failed to validate token", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims == nil || claims.UserID == "" || claims.Email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// JWT users have empty permissions = full access
		setContextValues(c, claims.UserID, claims.Email, []string{})
		c.Next()
	}
}
