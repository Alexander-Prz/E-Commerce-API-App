package db

import (
	"context"
	"fmt"

	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var Pool *pgxpool.Pool

func Connect() (*pgxpool.Pool, error) {
	ctx := context.Background()

	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Println("Connection failed!")
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		fmt.Println("Cannot access the database")
		return nil, err
	}

	return pool, nil
}
