package main

import (
	"fmt"
	"log"
	"os"

	"GameStoreAPI/external/abstractapi"
	"GameStoreAPI/external/midtrans"
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
	customerGamesRepo := repository.NewCustomerGamesRepository(pool)
	orderRepo := repository.NewOrderRepository(pool, customerGamesRepo)
	verifyRepo := repository.NewEmailVerificationRepository(pool)
	paymentRepo := repository.NewPaymentRepository(pool)
	snapClient := midtrans.NewSnapClient()

	// ======================
	// SERVICES
	// ======================
	authSvc := services.NewAuthService(authRepo, customerRepo, emailValidator, mailer, verifyRepo)
	devSvc := services.NewDeveloperService(devRepo)
	gameSvc := services.NewGameService(gameRepo, devRepo, customerRepo, customerGamesRepo)
	genreSvc := services.NewGenreService(genreRepo)
	gameGenreSvc := services.NewGameGenreService(gameGenreRepo, gameRepo, genreRepo)
	cartSvc := services.NewCartService(cartRepo, orderRepo, customerGamesRepo, authRepo, customerRepo)
	customerSvc := services.NewCustomerService(customerRepo, authRepo)
	customerGameSvc := services.NewCustomerGamesService(customerGamesRepo, cartRepo)
	orderSvc := services.NewOrderService(orderRepo)
	paymentService := services.NewPaymentService(paymentRepo, orderRepo, customerGamesRepo, cartRepo, snapClient)

	// ======================
	// ECHO
	// ======================
	e := echo.New()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	api := e.Group("/game-store")

	// ======================
	// ROUTES (ONLY REGISTRATION)
	// ======================
	registerAuthRoutes(api, authSvc)
	registerCustomerRoutes(api, customerSvc, authSvc)
	registerDeveloperRoutes(api, devSvc, authSvc)
	registerGameRoutes(api, gameSvc)
	registerGenreRoutes(api, genreSvc)
	registerGameGenreRoutes(api, gameGenreSvc)
	registerCartRoutes(api, cartSvc)
	registerOrderRoutes(api, orderSvc, cartSvc)
	registerCustomerGamesRoutes(api, customerGameSvc, customerSvc)
	registerPaymentRoutes(api, paymentService)

	// ======================
	// SERVER
	// ======================
	// Debug route listing
	for _, r := range e.Routes() {
		println(r.Method, r.Path)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
