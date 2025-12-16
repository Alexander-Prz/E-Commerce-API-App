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
	Qty    int   `json:"quantity"`
}

type updateCartRequest struct {
	Qty int `json:"quantity"`
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

	// POST /api/cart/checkout
	p.POST("/cart/checkout", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}
		orderID, err := cs.Checkout(c.Request().Context(), claims.AuthID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"orderid": orderID})
	})

	// ADD item
	p.POST("", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		req := new(addCartRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if req.Qty == 0 {
			req.Qty = 1
		}
		if err := cs.Add(c.Request().Context(), claims.AuthID, req.GameID, req.Qty); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusCreated, map[string]string{"message": "added"})
	})

	// UPDATE quantity
	p.PUT("/:gameid", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		gameID, _ := strconv.ParseInt(c.Param("gameid"), 10, 64)
		req := new(updateCartRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := cs.Update(c.Request().Context(), claims.AuthID, gameID, req.Qty); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "updated"})
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

	// CHECKOUT
	p.POST("/checkout", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		orderID, err := cs.Checkout(c.Request().Context(), claims.AuthID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"orderid": orderID})
	})
}
