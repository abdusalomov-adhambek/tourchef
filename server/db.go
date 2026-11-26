package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// connectDB opens a pgx pool and verifies connectivity before the HTTP server
// starts, so misconfiguration fails fast with a clear message.
func connectDB(ctx context.Context, dsn string) *pgxpool.Pool {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("invalid DATABASE_URL: %v", err)
	}
	poolCfg.MaxConns = 20
	poolCfg.MinConns = 1
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatalf("cannot create database pool: %v", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		log.Fatalf("cannot connect to PostgreSQL: %v", err)
	}
	log.Println("database connection established")
	return pool
}
