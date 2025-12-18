package main

import (
	"net/http"
	"strconv"

	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

/*
type createDeveloperRequest struct {
	DeveloperName string `json:"developername"`
	AuthID        *int64 `json:"authid,omitempty"`
}

type updateDeveloperRequest struct {
	DeveloperName string `json:"developername"`
	AuthID        *int64 `json:"authid,omitempty"`
}
*/
// registerDeveloperRoutes wires developer endpoints onto the provided group.
// - GET /developers            -> public list
// - GET /developers/:id        -> public get
// - POST /developers           -> admin only (create)
// - PUT /developers/:id        -> admin only (update)
// - DELETE /developers/:id     -> admin only (soft delete)
func registerDeveloperRoutes(g *echo.Group, devSvc *services.DeveloperService) {
	// public list
	g.GET("/developers", func(c echo.Context) error {
		list, err := devSvc.ListDevelopers(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, list)
	})

	// public get
	g.GET("/developers/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
		}
		dev, err := devSvc.GetDeveloper(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "developer not found"})
		}
		return c.JSON(http.StatusOK, dev)
	})
}
