package api

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// SupabaseClaims captures the common Supabase fields we care about.
type SupabaseClaims struct {
	Email string `json:"email,omitempty"`
	jwt.RegisteredClaims
}

var (
	supabaseJWTSecret []byte
	expectedIssuer    string
)

// InitSupabaseJWT loads the HMAC secret once at startup so requests only need to parse tokens.
func InitSupabaseJWT() error {
	if len(supabaseJWTSecret) > 0 {
		return nil
	}

	secret := strings.TrimSpace(os.Getenv("SUPABASE_JWT_SECRET"))
	if secret == "" {
		return fmt.Errorf("SUPABASE_JWT_SECRET must be set (Supabase Settings → API → JWT secret)")
	}

	supabaseJWTSecret = []byte(secret)
	expectedIssuer = strings.TrimSpace(os.Getenv("SUPABASE_JWT_ISSUER"))
	if expectedIssuer == "" {
		if projectID := strings.TrimSpace(os.Getenv("SUPABASE_PROJECT_ID")); projectID != "" {
			expectedIssuer = fmt.Sprintf("https://%s.supabase.co/auth/v1", projectID)
		}
	}

	log.Printf("Supabase JWT issuer constraint: %s", expectedIssuer)
	return nil
}

// VerifySupabaseJWT verifies signature + exp/nbf/iat + issuer and returns the claims.
//
// Expired tokens: jwt.ParseWithClaims enforces the standard registered claims, so if
// "exp" is in the past (or nbf is in the future) it returns an error and callers
// should respond with 401 Unauthorized. The backend never refreshes expired access
// tokens; the Supabase Swift client keeps a refresh token and automatically rotates
// the access token before expiry or after a 401. If refresh fails (offline, invalid
// refresh token, etc.) the client retries or asks the user to log in again.
func VerifySupabaseJWT(tokenStr string) (*SupabaseClaims, error) {
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return nil, fmt.Errorf("token is missing")
	}
	if len(supabaseJWTSecret) == 0 {
		return nil, fmt.Errorf("JWT verifier not initialized (call InitSupabaseJWT)")
	}

	claims := &SupabaseClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
		}
		return supabaseJWTSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, fmt.Errorf("parse or verify failed: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}
	if expectedIssuer != "" && claims.Issuer != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	return claims, nil
}

// SupabaseAuth returns a Fiber handler that enforces Authorization: Bearer <token>
// and exposes the claims via c.Locals("claims", *SupabaseClaims).
func SupabaseAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing Authorization header")
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid Authorization header")
		}

		claims, err := VerifySupabaseJWT(strings.TrimSpace(parts[1]))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		c.Locals("claims", claims)
		c.Locals(userIDContextKey, claims.Subject)
		return c.Next()
	}
}
