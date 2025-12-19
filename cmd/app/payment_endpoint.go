package main

import (
	"net/http"
	"strconv"

	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
)

func registerPaymentRoutes(g *echo.Group, ps *services.PaymentService) {
	p := g.Group("/payments")

	// ============================
	// MIDTRANS NOTIFICATION
	// (NO JWT, must be public)
	// ============================
	p.POST("/notification", func(c echo.Context) error {
		var payload map[string]interface{}
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusOK, echo.Map{
				"status": "ignored",
				"reason": "invalid payload",
			})
		}

		if err := ps.HandleMidtransNotification(
			c.Request().Context(),
			payload,
		); err != nil {
			// IMPORTANT:
			// Midtrans requires HTTP 200 or it will retry
			return c.JSON(http.StatusOK, echo.Map{
				"status": "ignored",
				"reason": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, echo.Map{
			"status": "ok",
		})
	})

	p.POST("/midtrans/webhook", func(c echo.Context) error {
		var payload map[string]interface{}
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid payload",
			})
		}

		if err := ps.HandleMidtransWebhook(
			c.Request().Context(),
			payload,
		); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, echo.Map{
			"status": "ok",
		})
	})

	// ============================
	// PAYMENT INITIATION
	// (JWT protected)
	// ============================
	p.Use(middleware.JWTMiddleware())

	p.POST("/:orderId", func(c echo.Context) error {
		cl := middleware.GetClaims(c)
		if cl == nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "unauthenticated",
			})
		}

		orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
		if err != nil || orderID <= 0 {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid order id",
			})
		}

		redirectURL, err := ps.CreateSnapPayment(
			c.Request().Context(),
			orderID,
			cl.AuthID,
		)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, echo.Map{
			"redirect_url": redirectURL,
		})
	})
}
