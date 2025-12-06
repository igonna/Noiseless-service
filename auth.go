package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultSupabaseIssuer = "http://127.0.0.1:54321/auth/v1"
	defaultJWKSURL        = "http://127.0.0.1:54321/auth/v1/keys"
)

type supabaseClaims struct {
	Email string `json:"email,omitempty"`
	jwt.RegisteredClaims
}

type supabaseUser struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
}

type jwksCache struct {
	url       string
	refreshIn time.Duration
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
}

func newJWKSCache(url string, refreshIn time.Duration) *jwksCache {
	return &jwksCache{url: url, refreshIn: refreshIn, keys: make(map[string]*rsa.PublicKey)}
}

func (c *jwksCache) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	valid := ok && time.Now().Before(c.expiresAt)
	c.mu.RUnlock()
	if valid {
		return key, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// Check again after obtaining write lock
	if key, ok := c.keys[kid]; ok && time.Now().Before(c.expiresAt) {
		return key, nil
	}
	if err := c.refreshLocked(ctx); err != nil {
		return nil, err
	}
	key, ok = c.keys[kid]
	if !ok {
		return nil, errors.New("key not found in JWKS")
	}
	return key, nil
}

func (c *jwksCache) refreshLocked(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("jwks fetch failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parsed, err := parseJWKS(body)
	if err != nil {
		return err
	}

	c.keys = parsed
	c.expiresAt = time.Now().Add(c.refreshIn)
	return nil
}

func parseJWKS(data []byte) (map[string]*rsa.PublicKey, error) {
	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(data, &jwks); err != nil {
		return nil, err
	}
	if len(jwks.Keys) == 0 {
		return nil, errors.New("jwks response contained no keys")
	}

	pubs := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, key := range jwks.Keys {
		if key.Kid == "" || strings.ToUpper(key.Kty) != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			return nil, err
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return nil, err
		}
		bigE := new(big.Int).SetBytes(eBytes)
		if bigE.BitLen() == 0 {
			return nil, errors.New("invalid jwks exponent")
		}
		pubs[key.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(bigE.Int64()),
		}
	}
	if len(pubs) == 0 {
		return nil, errors.New("no usable RSA keys in JWKS")
	}
	return pubs, nil
}

var (
	jwksOnce sync.Once
	jwksInst *jwksCache
	jwksErr  error
)

func jwksProvider() (*jwksCache, error) {
	jwksOnce.Do(func() {
		url := strings.TrimSpace(os.Getenv("SUPABASE_JWKS_URL"))
		if url == "" {
			url = defaultJWKSURL
		}
		jwksInst = newJWKSCache(url, 5*time.Minute)
	})
	return jwksInst, jwksErr
}

func authenticateSupabaseRequest(c *fiber.Ctx) (supabaseUser, error) {
	headers := c.Get("Authorization")
	if headers == "" {
		return supabaseUser{}, fiber.NewError(fiber.StatusUnauthorized, "missing Authorization header")
	}
	parts := strings.SplitN(headers, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return supabaseUser{}, fiber.NewError(fiber.StatusUnauthorized, "invalid Authorization header")
	}
	tokenString := strings.TrimSpace(parts[1])
	if tokenString == "" {
		return supabaseUser{}, fiber.NewError(fiber.StatusUnauthorized, "token missing")
	}

	cache, err := jwksProvider()
	if err != nil {
		return supabaseUser{}, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	claims := &supabaseClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("token missing kid")
		}
		return cache.getKey(context.Background(), kid)
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	if err != nil || !token.Valid {
		return supabaseUser{}, fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
	}

	issuer := strings.TrimSpace(os.Getenv("SUPABASE_JWT_ISSUER"))
	if issuer == "" {
		issuer = defaultSupabaseIssuer
	}
	if claims.Issuer != issuer {
		return supabaseUser{}, fiber.NewError(fiber.StatusUnauthorized, "invalid issuer")
	}
	if claims.Subject == "" {
		return supabaseUser{}, fiber.NewError(fiber.StatusUnauthorized, "token missing subject")
	}

	return supabaseUser{ID: claims.Subject, Email: claims.Email}, nil
}
