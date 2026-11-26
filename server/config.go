package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration. Every value can be overridden with
// an environment variable; nothing secret is hardcoded in the source tree.
type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   string
	Seed        bool
	CORSOrigin  string
	UploadDir   string
	SiteURL     string
}

func loadConfig() Config {
	// A local .env file is convenient in development but never required.
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:      getenv("APP_ENV", "dev"),
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tourchef?sslmode=disable"),
		JWTSecret:   getenv("JWT_SECRET", "tourchef-dev-secret-change-me"),
		JWTExpiry:   getenv("JWT_EXPIRY", "168h"), // 7 days
		Seed:        parseBool(getenv("SEED", "true")),
		CORSOrigin:  getenv("CORS_ORIGIN", "*"),
		UploadDir:   getenv("UPLOAD_DIR", "./uploads"),
		SiteURL:     strings.TrimRight(getenv("SITE_URL", "http://localhost:3000"), "/"),
	}

	if strings.TrimSpace(cfg.JWTSecret) == "" {
		log.Fatal("JWT_SECRET must not be empty")
	}
	if _, err := strconv.Atoi(cfg.Port); err != nil {
		log.Fatalf("invalid PORT %q: %v", cfg.Port, err)
	}
	return cfg
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}

func parseBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
