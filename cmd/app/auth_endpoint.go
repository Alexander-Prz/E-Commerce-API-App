package main

import (
	"net/http"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"` // admin-only when used via admin endpoints
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// registerPublic handles unauthenticated registration -> creates "user" role
func registerPublic(authSvc *services.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(registerRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		id, err := authSvc.RegisterPublic(c.Request().Context(), req.Email, req.Password)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusCreated, map[string]interface{}{"authid": id})
	}
}

// adminRegister allows admin to create admin/developer/user
func adminRegister(authSvc *services.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(registerRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		// role must be present and valid; AuthService will validate
		id, err := authSvc.RegisterByAdmin(c.Request().Context(), req.Email, req.Password, req.Role)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusCreated, map[string]interface{}{"authid": id})
	}
}

func loginHandler(authSvc *services.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(loginRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		user, err := authSvc.Login(c.Request().Context(), req.Email, req.Password)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}
		// generate JWT token
		claimsToken, err := middleware.GenerateToken(user.AuthID, user.Email, user.Role, 24)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not create token"})
		}
		// return token plus user info (without password)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"token":      claimsToken,
			"expires_in": 24 * 3600,
			"user": map[string]interface{}{
				"authid":     user.AuthID,
				"email":      user.Email,
				"role":       user.Role,
				"created_at": user.CreatedAt,
			},
		})
	}
}

// meHandler returns the authenticated user's info
func meHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}
		// return minimal info from token
		return c.JSON(http.StatusOK, map[string]interface{}{
			"authid": claims.AuthID,
			"email":  claims.Email,
			"role":   claims.Role,
			"exp":    claims.ExpiresAt,
		})
	}
}
