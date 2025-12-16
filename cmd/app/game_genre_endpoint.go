package main

import (
	"strconv"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

type gameGenreRequest struct {
	GenreID int64 `json:"genreid"`
}

func registerGameGenreRoutes(g *echo.Group, gs *services.GameGenreService) {
	// public get all genres of a game
	g.GET("/games/:id/genres", func(c echo.Context) error {
		idStr := c.Param("id")
		gameID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "invalid id"})
		}
		list, err := gs.ListGenres(c.Request().Context(), gameID)
		if err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, list)
	})

	// protected routes for admin and developer
	p := g.Group("")
	p.Use(middleware.JWTMiddleware())

	p.POST("/games/:id/genres", func(c echo.Context) error {
		// only admin or developer allowed
		cl := middleware.GetClaims(c)
		if cl == nil {
			return c.JSON(401, map[string]string{"error": "unauthenticated"})
		}
		if cl.Role != "admin" && cl.Role != "developer" {
			return c.JSON(403, map[string]string{"error": "forbidden"})
		}

		idStr := c.Param("id")
		gameID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "invalid id"})
		}

		req := new(gameGenreRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(400, map[string]string{"error": "invalid request"})
		}

		if err := gs.Add(c.Request().Context(), gameID, req.GenreID); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}
		return c.JSON(201, map[string]string{"message": "genre added"})
	})

	p.DELETE("/games/:id/genres/:genreid", func(c echo.Context) error {
		cl := middleware.GetClaims(c)
		if cl == nil {
			return c.JSON(401, map[string]string{"error": "unauthenticated"})
		}
		if cl.Role != "admin" && cl.Role != "developer" {
			return c.JSON(403, map[string]string{"error": "forbidden"})
		}

		gameID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		genreID, _ := strconv.ParseInt(c.Param("genreid"), 10, 64)

		if err := gs.Remove(c.Request().Context(), gameID, genreID); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, map[string]string{"message": "removed"})
	})
}
