# Supabase JWT Token Validation

This document describes the implementation of JWKS-based JWT token validation for Supabase authentication, replacing the previous HMAC-based validation.

## Overview

The `ValidateToken` method in the Supabase authentication provider has been modified to validate Supabase access tokens using JWKS (JSON Web Key Set) instead of requiring a symmetric JWT secret. This approach supports both ES256 and RS256 signing algorithms used by Supabase.

## Benefits

- **Security**: Uses asymmetric key validation instead of shared secrets
- **Algorithm Support**: Supports ES256/RS256 algorithms used by Supabase
- **Performance**: Eliminates HTTP calls for token validation, reducing latency
- **Reliability**: Removes dependency on Supabase API availability
- **Scalability**: Local validation scales better under high load
- **Cost**: Reduces API usage and potential rate limiting

## Implementation Details

### JWT Validation Process

1. **Parse JWT Token**: Uses the `github.com/golang-jwt/jwt/v5` library to parse the token
2. **JWKS Validation**: Validates the token signature using keys from Supabase's JWKS endpoint
3. **Check Expiration**: Automatically validates token expiration
4. **Validate Claims**: Ensures required claims are present and valid
5. **Extract User Info**: Extracts user ID, email, and phone from token claims

### Required Configuration

Add the publishable and secret keys to your configuration:

```yaml
supabase:
  url: "https://your-project.supabase.co"
  publishable_key: "pk_your_publishable_key"  # For client-side use
  secret_key: "sk_your_secret_key"            # For server-side use
  jwks_url: "https://your-project.supabase.co/auth/v1/.well-known/jwks.json"  # Optional, auto-derived
  jwt_secret: "your-supabase-jwt-secret"  # Optional, for HMAC fallback
```

Or via environment variables:

```bash
export CAYGNUS_SUPABASE_URL="https://your-project.supabase.co"
export CAYGNUS_SUPABASE_PUBLISHABLE_KEY="pk_your_publishable_key"
export CAYGNUS_SUPABASE_SECRET_KEY="sk_your_secret_key"
export CAYGNUS_SUPABASE_JWKS_URL="https://your-project.supabase.co/auth/v1/.well-known/jwks.json"  # Optional
export CAYGNUS_SUPABASE_JWT_SECRET="your-supabase-jwt-secret"  # Optional
```

### JWKS URL Derivation

If `jwks_url` is not provided, it will be automatically derived from the Supabase URL by appending `/auth/v1/.well-known/jwks.json`.

### Finding Your Configuration Values

1. Go to your Supabase Dashboard
2. Navigate to Settings > API
3. Copy the following values:
   - **Project URL**: Use as `supabase.url`
   - **Publishable Key**: Use as `supabase.publishable_key` (for client-side use)
   - **Secret Key**: Use as `supabase.secret_key` (for server-side use)
   - **JWT Secret**: Use as `supabase.jwt_secret` (optional, for HMAC fallback)
4. The JWKS URL will be automatically derived as `{supabase.url}/auth/v1/.well-known/jwks.json`

## Token Validation Rules

The implementation validates the following:

### Required Claims

- `sub`: User ID (must be present and non-empty)
- `aud`: Audience (must be "authenticated" for user tokens)
- `role`: Role (must be "authenticated" for user tokens)
- `exp`: Expiration (automatically validated by JWT library)

### Optional Claims

- `email`: User email address
- `phone`: User phone number
- `iat`: Issued at timestamp

### Supported Algorithms

- **ES256**: ECDSA using P-256 and SHA-256 (Supabase default)
- **RS256**: RSA using SHA-256
- **HS256**: HMAC SHA-256 (fallback only, requires JWT secret)

## Usage Example

```go
// The ValidateToken method is called automatically by your auth middleware
claims, err := authProvider.ValidateToken(ctx, accessToken)
if err != nil {
    // Token is invalid
    return err
}

// Use the validated claims
userID := claims.UserID
email := claims.Email
phone := claims.Phone
```

## Error Handling

The implementation returns specific errors for different validation failures:

- `invalid token: signature is invalid` - Wrong JWT secret or tampered token
- `token is invalid` - General token validation failure
- `token is expired` - Token has expired
- `invalid token audience` - Wrong audience claim
- `invalid token role` - Wrong role claim
- `token missing user ID` - Missing or empty sub claim

## Testing

Comprehensive tests are included in `internal/auth/supabase_test.go`:

```bash
go test ./internal/auth -v
```

Test scenarios include:

- Valid JWT token validation
- Invalid JWT secret handling
- Missing user ID detection
- Invalid audience rejection
- Expired token rejection

## Migration from HTTP Validation

### Before (HTTP-based)

```go
func (p *supabaseProvider) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
    claims, err := p.supabase.Admin.GetUser(ctx, token)
    if err != nil {
        return nil, err
    }
    return &auth.Claims{
        UserID: claims.ID,
    }, nil
}
```

### After (JWT-based)

```go
func (p *supabaseProvider) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
    // Parse and validate JWT locally
    parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method and return secret
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(p.cfg.Supabase.JWTSecret), nil
    })

    // Validate and extract claims...
}
```

## Security Considerations

1. **JWT Secret Security**: Keep your JWT secret secure and never expose it in client-side code
2. **Token Expiration**: JWT tokens have built-in expiration validation
3. **Signature Verification**: All tokens are cryptographically verified
4. **Claim Validation**: Required claims are validated to ensure token integrity

## Troubleshooting

### Common Issues

1. **"signature is invalid"**

   - Check that your JWT secret is correct
   - Ensure the secret matches your Supabase project

2. **"invalid token audience"**

   - Ensure you're using user access tokens, not anon keys
   - Check that the token was issued for authenticated users

3. **"token is expired"**
   - Token has expired and needs to be refreshed
   - Check your token refresh logic

### Debug Logging

Enable debug logging to see detailed validation information:

```yaml
logging:
  level: "debug"
```

## Dependencies

- `github.com/golang-jwt/jwt/v5`: JWT parsing and validation library

## References

- [Supabase JWT Documentation](https://supabase.com/docs/guides/auth/jwts)
- [JWT.io Debugger](https://jwt.io/) - For testing and debugging JWT tokens
- [golang-jwt/jwt Documentation](https://github.com/golang-jwt/jwt)
