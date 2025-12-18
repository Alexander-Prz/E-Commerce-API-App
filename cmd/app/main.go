package main

import (
	"fmt"
	"log"
	"os"

	"GameStoreAPI/external/abstractapi"
	"GameStoreAPI/external/resend"

	"GameStoreAPI/internal/db"
	"GameStoreAPI/internal/repository"
	"GameStoreAPI/internal/services"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	fmt.Println("MEONG!")
	// ======================
	// INFRA
	// ======================
	pool, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// ======================
	// EXTERNALS
	// ======================
	useExternal := os.Getenv("USE_EMAIL_REPUTATION") == "true"

	var emailValidator services.EmailValidator
	if useExternal {
		emailValidator, err = abstractapi.NewAbstractReputationValidator()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		emailValidator = services.NewLocalValidator()
	}

	mailer, err := resend.NewResendMailer("PrzGameStore<onboarding@resend.dev>")
	if err != nil {
		log.Fatal(err)
	}

	// ======================
	// REPOSITORIES
	// ======================
	authRepo := repository.NewAuthRepository(pool)
	devRepo := repository.NewDeveloperRepository(pool)
	gameRepo := repository.NewGameRepository(pool)
	genreRepo := repository.NewGenreRepository(pool)
	gameGenreRepo := repository.NewGameGenreRepository(pool)
	cartRepo := repository.NewCartRepository(pool)
	customerRepo := repository.NewCustomerRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	customerGamesRepo := repository.NewCustomerGamesRepository(pool)
	verifyRepo := repository.NewEmailVerificationRepository(pool)

	// ======================
	// SERVICES
	// ======================
	authSvc := services.NewAuthService(authRepo, customerRepo, emailValidator, mailer, verifyRepo)
	devSvc := services.NewDeveloperService(devRepo)
	gameSvc := services.NewGameService(gameRepo, devRepo)
	genreSvc := services.NewGenreService(genreRepo)
	gameGenreSvc := services.NewGameGenreService(gameGenreRepo, gameRepo, genreRepo)
	cartSvc := services.NewCartService(cartRepo, orderRepo, customerGamesRepo, authRepo, customerRepo)
	customerSvc := services.NewCustomerService(customerRepo, authRepo)
	orderSvc := services.NewOrderService(orderRepo)
	customerGameSvc := services.NewCustomerGamesService(customerGamesRepo, cartRepo)

	// ======================
	// ECHO
	// ======================
	e := echo.New()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	api := e.Group("/api")

	// ======================
	// ROUTES (ONLY REGISTRATION)
	// ======================
	registerAuthRoutes(api, authSvc)
	registerCustomerRoutes(api, customerSvc)
	registerDeveloperRoutes(api, devSvc)
	registerGameRoutes(api, gameSvc)
	registerGenreRoutes(api, genreSvc)
	registerGameGenreRoutes(api, gameGenreSvc)
	registerCartRoutes(api, cartSvc)
	registerOrderRoutes(api, orderSvc)
	registerCustomerGamesRoutes(api, customerGameSvc, customerSvc)

	// ======================
	// SERVER
	// ======================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
