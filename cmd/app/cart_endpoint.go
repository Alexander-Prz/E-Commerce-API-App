package main

import (
	"net/http"
	"strconv"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

type addCartRequest struct {
	GameID int64 `json:"gameid"`
}

func registerCartRoutes(g *echo.Group, cs *services.CartService) {
	p := g.Group("/cart")
	p.Use(middleware.JWTMiddleware())

	// GET cart
	p.GET("", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		cart, err := cs.Get(c.Request().Context(), claims.AuthID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, cart)
	})

	// ADD item
	p.POST("", func(c echo.Context) error {
		claims := middleware.GetClaims(c)

		req := new(addCartRequest)
		if err := c.Bind(req); err != nil || req.GameID == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request",
			})
		}

		if err := cs.Add(
			c.Request().Context(),
			claims.AuthID,
			req.GameID,
		); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusCreated, map[string]string{
			"message": "added to cart",
		})
	})

	// REMOVE item
	p.DELETE("/:gameid", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		gameID, _ := strconv.ParseInt(c.Param("gameid"), 10, 64)
		if err := cs.Remove(c.Request().Context(), claims.AuthID, gameID); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "removed"})
	})

	// CLEAR cart
	p.DELETE("", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if err := cs.Clear(c.Request().Context(), claims.AuthID); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "cleared"})
	})

}
