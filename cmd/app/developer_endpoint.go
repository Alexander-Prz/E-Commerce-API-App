package main

import (
	"net/http"
	"strconv"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

// request payloads
type createDeveloperRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	DeveloperName string `json:"developername"`
}

type updateDeveloperRequest struct {
	DeveloperName string `json:"developername"`
}

// registerDeveloperRoutes wires developer endpoints
func registerDeveloperRoutes(api *echo.Group, devSvc *services.DeveloperService, authSvc *services.AuthService) {

	// =====================
	// Public developer APIs
	// =====================

	// GET /api/developers
	api.GET("/developers", func(c echo.Context) error {
		list, err := devSvc.ListDevelopers(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, list)
	})

	// GET /api/developers/:id
	api.GET("/developers/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid id",
			})
		}

		dev, err := devSvc.GetDeveloper(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error": "developer not found",
			})
		}
		return c.JSON(http.StatusOK, dev)
	})

	// ==========================
	// Admin developer management
	// ==========================

	admin := api.Group("/admin")
	admin.Use(middleware.JWTMiddleware())
	admin.Use(middleware.AdminOnly)

	// POST /api/admin/developers
	admin.POST("/developers", func(c echo.Context) error {
		req := new(createDeveloperRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid request",
			})
		}

		// 1️⃣ Create auth account (role = developer)
		authID, err := authSvc.RegisterByAdmin(
			c.Request().Context(),
			req.Email,
			req.Password,
			"developer",
		)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		// 2️⃣ Create developer profile
		devID, err := devSvc.CreateDeveloper(
			c.Request().Context(),
			req.DeveloperName,
			&authID,
		)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusCreated, echo.Map{
			"authid":      authID,
			"developerid": devID,
		})
	})

	// PUT /api/admin/developers/:id
	admin.PUT("/developers/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		devID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid id",
			})
		}

		req := new(updateDeveloperRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid request",
			})
		}

		// 🔍 Get developer to ensure it exists and fetch authid
		dev, err := devSvc.GetDeveloper(c.Request().Context(), devID)
		if err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error": "developer not found",
			})
		}

		// ✅ Update developer profile ONLY
		if err := devSvc.UpdateDeveloper(
			c.Request().Context(),
			dev.DeveloperID,
			req.DeveloperName,
			dev.AuthID, // unchanged
		); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "developer updated",
		})
	})

	// DELETE /api/admin/developers/:id
	admin.DELETE("/developers/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		devID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid id",
			})
		}

		// 🔍 Load developer to get authid
		dev, err := devSvc.GetDeveloper(c.Request().Context(), devID)
		if err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error": "developer not found",
			})
		}

		// 1️⃣ Soft-delete developer
		if err := devSvc.DeleteDeveloper(c.Request().Context(), devID); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		// 2️⃣ Ban auth account (if linked)
		if dev.AuthID != nil {
			_ = authSvc.BanUser(c.Request().Context(), *dev.AuthID)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "developer deleted and auth banned",
		})
	})
}
