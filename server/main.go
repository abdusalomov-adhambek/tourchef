package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	cfg Config
	db  *pgxpool.Pool
)

func main() {
	cfg = loadConfig()
	ctx := context.Background()

	db = connectDB(ctx, cfg.DatabaseURL)
	defer db.Close()

	if err := migrate(ctx, db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	if err := seedIfEnabled(ctx, db, cfg.Seed); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	app := newApp()
	log.Printf("TourChef API listening on :%s (env=%s, seed=%v)", cfg.Port, cfg.AppEnv, cfg.Seed)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
