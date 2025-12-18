package main

import (
	"net/http"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

func registerOrderRoutes(g *echo.Group, os *services.OrderService) {

	p := g.Group("/orders")
	p.Use(middleware.JWTMiddleware())

	p.POST("/checkout", func(c echo.Context) error {
		cl := middleware.GetClaims(c)
		if cl == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "unauthenticated",
			})
		}

		customerID := cl.AuthID

		orderID, err := os.Checkout(c.Request().Context(), customerID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"order_id": orderID,
			"status":   "pending",
		})
	})
}
