package main

import (
	"net/http"
	"strconv"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

func registerOrderRoutes(g *echo.Group, os *services.OrderService, cs *services.CartService) {

	p := g.Group("/orders")
	p.Use(middleware.JWTMiddleware())

	p.POST("/checkout", func(c echo.Context) error {
		cl := middleware.GetClaims(c)
		if cl == nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "unauthenticated",
			})
		}

		orderID, err := cs.Checkout(
			c.Request().Context(),
			cl.AuthID,
		)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusCreated, echo.Map{
			"order_id": orderID,
			"status":   "pending_payment",
		})
	})

	p.GET("/:id/status", func(c echo.Context) error {
		orderID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		order, err := os.Repo.GetOrderByID(c.Request().Context(), orderID)
		if err != nil {
			return c.JSON(404, echo.Map{"error": "not found"})
		}

		return c.JSON(200, echo.Map{
			"order_id": order.OrderID,
			"status":   order.OrderStatus,
		})
	})

}
