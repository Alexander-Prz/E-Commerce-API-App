package main

import (
	"net/http"
	"strconv"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

// user update payload
type updateCustomerRequest struct {
	Fullname *string `json:"fullname,omitempty"`
	Address  *string `json:"address,omitempty"`
	Phone    *string `json:"phone,omitempty"`
}

// register customer routes (user self routes + admin user management)
func registerCustomerRoutes(api *echo.Group, cs *services.CustomerService, as *services.AuthService) {
	// User routes (require JWT)
	userGrp := api.Group("/customers")
	userGrp.Use(middleware.JWTMiddleware())

	// GET /api/customers/me
	userGrp.GET("/me", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}
		cust, err := cs.GetByAuthID(c.Request().Context(), claims.AuthID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "customer not found"})
		}
		return c.JSON(http.StatusOK, cust)
	})

	// PUT /api/customers/me
	userGrp.PUT("/me", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}
		cust, err := cs.GetByAuthID(c.Request().Context(), claims.AuthID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "customer not found"})
		}
		req := new(updateCustomerRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if req.Fullname == nil && req.Address == nil && req.Phone == nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "nothing to update"})
		}
		if err := cs.UpdateSelf(c.Request().Context(), cust.CustomerID, req.Fullname, req.Address, req.Phone); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "updated"})
	})

	// Admin management group
	admin := api.Group("/admin")
	admin.Use(middleware.JWTMiddleware())
	admin.Use(middleware.AdminOnly)

	// LIST all users with role='user'
	admin.GET("/users", func(c echo.Context) error {
		users, err := cs.Users.ListUsersOnly(c.Request().Context())
		if err != nil {
			return c.JSON(500, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, users)
	})

	// GET detail for a specific user with role='user'
	admin.GET("/users/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "invalid id"})
		}

		user, err := cs.Users.GetUserOnlyByID(c.Request().Context(), id)
		if err != nil {
			return c.JSON(404, map[string]string{"error": "user not found"})
		}

		return c.JSON(200, user)
	})

	// BAN user (soft delete)
	admin.POST("/users/:id/ban", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "invalid id"})
		}

		if err := as.BanUser(c.Request().Context(), id); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}

		return c.JSON(200, map[string]string{"message": "user banned"})
	})

	admin.POST("/users/:id/unban", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "invalid id"})
		}

		if err := as.UnBanUser(c.Request().Context(), id); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}

		return c.JSON(200, map[string]string{"message": "user ban was lifted"})
	})
}
