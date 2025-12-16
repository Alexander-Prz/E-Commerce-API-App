package main

import (
	"net/http"
	"strconv"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

func registerCustomerGamesRoutes(g *echo.Group, cgSvc *services.CustomerGamesService, cs *services.CustomerService) {

	usr := g.Group("/customers/me")
	usr.Use(middleware.JWTMiddleware())

	// GET /api/customers/me/games
	usr.GET("/games", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}

		cust, err := cs.GetByAuthID(c.Request().Context(), claims.AuthID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "customer not found"})
		}

		owned, err := cgSvc.ListOwned(c.Request().Context(), cust.CustomerID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, owned)
	})

	// GET /api/customers/me/orders
	usr.GET("/orders", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}

		cust, err := cs.GetByAuthID(c.Request().Context(), claims.AuthID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "customer not found"})
		}

		orders, err := cgSvc.ListOrders(c.Request().Context(), cust.CustomerID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, orders)
	})

	// GET /api/customers/me/orders/:id
	usr.GET("/orders/:id", func(c echo.Context) error {
		claims := middleware.GetClaims(c)
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		}

		cust, err := cs.GetByAuthID(c.Request().Context(), claims.AuthID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "customer not found"})
		}

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		}

		order, items, err := cgSvc.OrderDetails(c.Request().Context(), cust.CustomerID, id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"order": order,
			"items": items,
		})
	})

	admin := g.Group("/admin")
	admin.Use(middleware.JWTMiddleware())
	admin.Use(middleware.AdminOnly)

	admin.GET("/orders", func(c echo.Context) error {
		orders, err := cgSvc.ListAllOrders(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, orders)
	})

	admin.GET("/orders/:id", func(c echo.Context) error {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
		}

		o, items, err := cgSvc.GetOrderDetailsAdmin(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"order": o,
			"items": items,
		})
	})

}
