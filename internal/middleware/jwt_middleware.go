package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Claims defines JWT payload structure
type Claims struct {
	AuthID int64  `json:"authid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-please-change"
	}
	jwtSecret = []byte(secret)
}

// GenerateToken creates a signed token for the given user details and expiry (in hours)
func GenerateToken(authid int64, email, role string, hours int) (string, error) {
	claims := &Claims{
		AuthID: authid,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(hours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "yesh-api",
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(jwtSecret)
}

// JWTMiddleware returns an Echo middleware that validates token and sets "user" context
func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			}
			parts := strings.Fields(auth)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header"})
			}
			tokenString := parts[1]
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			}
			claims, ok := token.Claims.(*Claims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			}
			// attach claims to context
			c.Set("auth_claims", claims)
			return next(c)
		}
	}
}

// Helper to extract claims
func GetClaims(c echo.Context) *Claims {
	v := c.Get("auth_claims")
	if v == nil {
		return nil
	}
	if cl, ok := v.(*Claims); ok {
		return cl
	}
	return nil
}

// TryGetClaimsFromAuthHeader checks Authorization header and parses token if present.
// Returns claims or nil (no error). If token is present but invalid, returns nil.
func TryGetClaimsFromAuthHeader(c echo.Context) *Claims {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return nil
	}
	parts := strings.Fields(auth)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil
	}
	tokenString := parts[1]
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// AdminOnly middleware requires role == admin
func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims := GetClaims(c)
		if claims == nil || claims.Role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "admin role required"})
		}
		return next(c)
	}
}
