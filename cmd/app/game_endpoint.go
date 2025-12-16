package main

import (
	"net/http"
	"strconv"
	"time"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/model"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

// request payloads
type createGameRequest struct {
	DeveloperID int64   `json:"developerid"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	ReleaseDate string  `json:"releasedate,omitempty"` // YYYY-MM-DD expected
}

type updateGameRequest struct {
	DeveloperID int64   `json:"developerid"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	ReleaseDate string  `json:"releasedate,omitempty"`
}

// registerGameRoutes mounts game endpoints to the provided group.
// Public:
//
//	GET /games         -> list (pagination via ?limit=&offset=)
//	GET /games/:id     -> get
//
// Protected (admin OR developer):
//
//	POST /games        -> create
//	PUT /games/:id     -> update
//	DELETE /games/:id  -> soft delete
func registerGameRoutes(g *echo.Group, gs *services.GameService) {
	// public list
	g.GET("/games", func(c echo.Context) error {
		// If request includes Authorization header and it belongs to a developer,
		// disallow this endpoint and instruct to use developer-only endpoint.
		claims := middleware.TryGetClaimsFromAuthHeader(c)
		if claims != nil && claims.Role == "developer" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "developers must use /api/developer/games to view their games"})
		}

		limitStr := c.QueryParam("limit")
		offsetStr := c.QueryParam("offset")
		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)
		list, err := gs.ListGames(c.Request().Context(), limit, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, list)
	})

	// public get
	g.GET("/games/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
		}
		game, err := gs.GetGame(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "game not found"})
		}
		return c.JSON(http.StatusOK, game)
	})

	// developer-only "my games" endpoint (protected)
	devGroup := g.Group("/developer")
	devGroup.Use(middleware.JWTMiddleware())
	devGroup.GET("/games", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}
		if claims.Role != "developer" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "developer role required"})
		}
		// find developer by authid
		dev, err := gs.DeveloperRepo.GetByAuthID(c.Request().Context(), claims.AuthID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "developer record not found for this account"})
		}
		limitStr := c.QueryParam("limit")
		offsetStr := c.QueryParam("offset")
		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)
		list, err := gs.ListGamesByDeveloper(c.Request().Context(), dev.DeveloperID, limit, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, list)
	})

	// protected group for create/update/delete (requires JWT)
	protected := g.Group("")
	protected.Use(middleware.JWTMiddleware())

	// create - admin or developer
	protected.POST("/games", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}
		// bind request
		req := new(createGameRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		// parse release date if provided
		var rd *time.Time
		if req.ReleaseDate != "" {
			t, err := time.Parse("2006-01-02", req.ReleaseDate)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid release date (use YYYY-MM-DD)"})
			}
			rd = &t
		}
		// authorization: admin can create any developerid; developer can only create games for their own developer record
		if claims.Role == "developer" {
			// ensure developer's authid matches the developer record
			dev, err := gs.DeveloperRepo.GetByID(c.Request().Context(), req.DeveloperID)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "developer not found"})
			}
			if dev.AuthID == nil || *dev.AuthID != claims.AuthID {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "developers can only create games for their own developer record"})
			}
		} else if claims.Role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only admin or developer roles can create games"})
		}
		game := &model.Game{
			DeveloperID: req.DeveloperID,
			Title:       req.Title,
			Price:       req.Price,
			ReleaseDate: rd,
		}
		id, err := gs.CreateGame(c.Request().Context(), game)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusCreated, map[string]interface{}{"gameid": id})
	})

	// update
	protected.PUT("/games/:id", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
		}
		req := new(updateGameRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		var rd *time.Time
		if req.ReleaseDate != "" {
			t, err := time.Parse("2006-01-02", req.ReleaseDate)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid release date (use YYYY-MM-DD)"})
			}
			rd = &t
		}
		// fetch existing game to check ownership
		existing, err := gs.GetGame(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "game not found"})
		}
		if claims.Role == "developer" {
			// ensure the developer making the request owns the game
			dev, err := gs.DeveloperRepo.GetByID(c.Request().Context(), existing.DeveloperID)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "developer not found"})
			}
			if dev.AuthID == nil || *dev.AuthID != claims.AuthID {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "developers can only update their own games"})
			}
		} else if claims.Role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only admin or developer roles can update games"})
		}
		update := &model.Game{
			GameID:      id,
			DeveloperID: req.DeveloperID,
			Title:       req.Title,
			Price:       req.Price,
			ReleaseDate: rd,
		}
		if err := gs.UpdateGame(c.Request().Context(), update); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "updated"})
	})

	// delete
	protected.DELETE("/games/:id", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
		}
		existing, err := gs.GetGame(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "game not found"})
		}
		if claims.Role == "developer" {
			dev, err := gs.DeveloperRepo.GetByID(c.Request().Context(), existing.DeveloperID)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "developer not found"})
			}
			if dev.AuthID == nil || *dev.AuthID != claims.AuthID {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "developers can only delete their own games"})
			}
		} else if claims.Role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only admin or developer roles can delete games"})
		}
		if err := gs.DeleteGame(c.Request().Context(), id); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
	})
}
