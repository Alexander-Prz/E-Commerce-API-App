package main

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"GameStoreAPI/internal/services"
)

func verifyEmailHandler(authSvc *services.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")
		if token == "" {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "token required",
			})
		}

		if err := authSvc.VerifyEmail(
			c.Request().Context(),
			token,
		); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "failed to verify email",
			})
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "email verified",
		})
	}
}
