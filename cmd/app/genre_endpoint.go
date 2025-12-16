package main

import (
	"strconv"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

type createGenreRequest struct {
	GenreName string `json:"genrename"`
}

type updateGenreRequest struct {
	GenreName string `json:"genrename"`
}

func registerGenreRoutes(g *echo.Group, gs *services.GenreService) {

	// PUBLIC — list genres
	g.GET("/genres", func(c echo.Context) error {
		list, err := gs.List(c.Request().Context())
		if err != nil {
			return c.JSON(500, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, list)
	})

	// PUBLIC — get genre
	g.GET("/genres/:id", func(c echo.Context) error {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "invalid id"})
		}
		genre, err := gs.Get(c.Request().Context(), id)
		if err != nil {
			return c.JSON(404, map[string]string{"error": "genre not found"})
		}
		return c.JSON(200, genre)
	})

	// PROTECTED — admin only write operations
	admin := g.Group("/genres")
	admin.Use(middleware.JWTMiddleware())
	admin.Use(middleware.AdminOnly)

	// CREATE
	admin.POST("", func(c echo.Context) error {
		req := new(createGenreRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(400, map[string]string{"error": "invalid request"})
		}
		id, err := gs.Create(c.Request().Context(), req.GenreName)
		if err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}
		return c.JSON(201, map[string]interface{}{"genreid": id})
	})

	// UPDATE
	admin.PUT("/:id", func(c echo.Context) error {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		req := new(updateGenreRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(400, map[string]string{"error": "invalid request"})
		}
		if err := gs.Update(c.Request().Context(), id, req.GenreName); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, map[string]string{"message": "updated"})
	})

	// DELETE
	admin.DELETE("/:id", func(c echo.Context) error {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		if err := gs.Delete(c.Request().Context(), id); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, map[string]string{"message": "deleted"})
	})
}
