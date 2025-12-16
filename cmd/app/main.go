package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"GameStoreAPI/internal/db"
	"GameStoreAPI/internal/middleware"
	"GameStoreAPI/internal/repository"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	// DB connect
	pool, err := db.Connect()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("cannot ping db: %v", err)
	}

	// repositories
	authRepo := repository.NewAuthRepository(pool)
	devRepo := repository.NewDeveloperRepository(pool)
	gameRepo := repository.NewGameRepository(pool)
	genreRepo := repository.NewGenreRepository(pool)
	gameGenreRepo := repository.NewGameGenreRepository(pool)
	cartRepo := repository.NewCartRepository(pool)
	customerRepo := repository.NewCustomerRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	customerGamesRepo := repository.NewCustomerGamesRepository(pool)

	// services
	authSvc := services.NewAuthService(authRepo, customerRepo)
	devSvc := services.NewDeveloperService(devRepo)
	gameSvc := services.NewGameService(gameRepo, devRepo)
	genreSvc := services.NewGenreService(genreRepo)
	gameGenreSvc := services.NewGameGenreService(gameGenreRepo, gameRepo, genreRepo)
	cartSvc := services.NewCartService(cartRepo, orderRepo, customerGamesRepo, authRepo, customerRepo)
	customerSvc := services.NewCustomerService(customerRepo, authRepo)
	customerGameSvc := services.NewCustomerGamesService(customerGamesRepo, cartRepo)

	// Echo
	e := echo.New()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	api := e.Group("/api")

	// ======================
	// AUTH ENDPOINTS
	// ======================
	api.POST("/auth/register", registerPublic(authSvc))
	api.POST("/auth/login", loginHandler(authSvc))

	authGroup := api.Group("/auth")
	authGroup.Use(middleware.JWTMiddleware())
	authGroup.GET("/me", meHandler())

	registerCustomerRoutes(api, customerSvc)
	registerDeveloperRoutes(api, devSvc)
	registerGameRoutes(api, gameSvc)
	registerGenreRoutes(api, genreSvc)
	registerGameGenreRoutes(api, gameGenreSvc)
	registerCartRoutes(api, cartSvc)
	registerCustomerGamesRoutes(api, customerGameSvc, customerSvc)

	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTMiddleware())
	adminGroup.Use(middleware.AdminOnly)

	// Admin can register admin/dev accounts
	adminGroup.POST("/auth/register", adminRegister(authSvc))

	// Admin CRUD developer
	adminGroup.POST("/developers", func(c echo.Context) error {
		req := new(struct {
			DeveloperName string `json:"developername"`
			AuthID        *int64 `json:"authid"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(400, map[string]string{"error": "invalid request"})
		}
		id, err := devSvc.CreateDeveloper(c.Request().Context(), req.DeveloperName, req.AuthID)
		if err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}
		return c.JSON(201, map[string]interface{}{"developerid": id})
	})

	adminGroup.PUT("/developers/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "invalid id"})
		}

		req := new(struct {
			DeveloperName *string `json:"developername"`
			AuthID        *int64  `json:"authid"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(400, map[string]string{"error": "invalid request"})
		}

		// Enforce required developer name (since service requires it)
		if req.DeveloperName == nil || strings.TrimSpace(*req.DeveloperName) == "" {
			return c.JSON(400, map[string]string{"error": "developer name is required"})
		}

		// Extract the value (safe because we validated above)
		devName := strings.TrimSpace(*req.DeveloperName)

		if err := devSvc.UpdateDeveloper(c.Request().Context(), id, devName, req.AuthID); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}

		return c.JSON(200, map[string]string{"message": "updated"})
	})

	adminGroup.DELETE("/developers/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "invalid id"})
		}
		if err := devSvc.DeleteDeveloper(c.Request().Context(), id); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, map[string]string{"message": "deleted"})
	})

	// ======================
	// START SERVER
	// ======================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("starting server on :%s\n", port)

	// Debug route listing
	for _, r := range e.Routes() {
		println(r.Method, r.Path)
	}

	e.Logger.Fatal(e.Start(":" + port))
}
