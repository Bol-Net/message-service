package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Custom claims to include role and name information
type CustomClaims struct {
	Role string `json:"role"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}

func LoadPublicKey() (*rsa.PublicKey, error) {
	pub := os.Getenv("JWT_PUBLIC_KEY")
	if pub == "" {
		return nil, fmt.Errorf("missing JWT_PUBLIC_KEY")
	}
	return jwt.ParseRSAPublicKeyFromPEM([]byte(pub))
}

func VerifyToken(tokenStr string, pubKey *rsa.PublicKey) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method is RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// Extract Bearer token from Authorization header
func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

// Extract token from query parameter (for WebSocket)
func ExtractQueryToken(r *http.Request) (string, error) {
	token := r.URL.Query().Get("token")
	if token == "" {
		return "", fmt.Errorf("missing token query parameter")
	}
	return token, nil
}

// JWT Middleware for HTTP endpoints
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load public key
		pubKey, err := LoadPublicKey()
		if err != nil {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Extract token
		tokenStr, err := ExtractBearerToken(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Verify token
		claims, err := VerifyToken(tokenStr, pubKey)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", claims.Subject)
		ctx = context.WithValue(ctx, "user_role", claims.Role)

		// Call next handler with context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Get user ID from context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in context")
	}
	return userID, nil
}

// Get user role from context
func GetUserRoleFromContext(ctx context.Context) (string, error) {
	role, ok := ctx.Value("user_role").(string)
	if !ok {
		return "", fmt.Errorf("user_role not found in context")
	}
	return role, nil
}
