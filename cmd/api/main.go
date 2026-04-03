package main

import (
	"log"
	"time"

	"github.com/Petroviiic/finance_api_backend/internal/auth"
	"github.com/Petroviiic/finance_api_backend/internal/db"
	"github.com/Petroviiic/finance_api_backend/internal/env"
	"github.com/Petroviiic/finance_api_backend/internal/ratelimiter"
	"github.com/Petroviiic/finance_api_backend/internal/storage"
	"github.com/joho/godotenv"
)

// @title Finance API Backend
// @version 1.0
// @host localhost:3000
// @BasePath /v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Type "Bearer <your-jwt-token>"
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading env file: %v", err)
	}
	cfg := Config{
		addr:      env.GetString("ADDR", ":3000"),
		isProdEnv: env.GetBool("IS_PROD_ENV", false),
		dashboard: dashboardConfig{
			NumberOfRecentRecords: env.GetInt("DASHBOARD_RECORD_NUMBER", 5),
		},
		dbConfig: DBConfig{
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
			dbAddr:       env.GetString("DB_ADDR", "postgresql://user:user123@localhost:5432/finance_api_db?sslmode=disable"),
		},
		auth: authConfig{
			secret:  env.GetString("AUTH_TOKEN_SECRET", "test"),
			expDate: time.Hour * 24 * 3,
			iss:     env.GetString("AUTH_TOKEN_ISSUER", "admin"),
		},
		ratelimiter: rateLimiterConfig{
			authFixedWindow: fixedWindowLimiterConfig{
				limit:  15,
				window: 3 * time.Minute,
			},
			apiFixedWindow: fixedWindowLimiterConfig{
				limit:  10,
				window: 1 * time.Minute,
			},
			tokenBucket: tokenBucketLimiterConfig{
				limit:           15,
				tokensPerMinute: 5,
			},
		},
	}

	db, err := db.NewDb(cfg.dbConfig.dbAddr, cfg.dbConfig.maxIdleConns, cfg.dbConfig.maxOpenConns, cfg.dbConfig.maxIdleTime)
	if err != nil {
		log.Panic("error connecting to db")
		return
	}
	if err := runMigrations(cfg.dbConfig.dbAddr); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	storage := storage.NewStorage(db)

	authenticator := auth.NewJWTAuthenticator(cfg.auth.secret, cfg.auth.iss, cfg.auth.iss)

	authFixedWindowLimiter := ratelimiter.NewFixedWindowLimiter(cfg.ratelimiter.authFixedWindow.limit, cfg.ratelimiter.authFixedWindow.window)
	authFixedWindowLimiter.Cleanup()

	apiFixedWindowLimiter := ratelimiter.NewFixedWindowLimiter(cfg.ratelimiter.apiFixedWindow.limit, cfg.ratelimiter.apiFixedWindow.window)
	apiFixedWindowLimiter.Cleanup()

	apiTokenBuckerLimiter := ratelimiter.NewTokenBuckerRatelimiter(cfg.ratelimiter.tokenBucket.limit, cfg.ratelimiter.tokenBucket.tokensPerMinute)
	apiTokenBuckerLimiter.Cleanup()

	app := &Application{
		config:        cfg,
		db:            db,
		storage:       storage,
		authenticator: authenticator,
		rateLimiters: rateLimiters{
			authFixedWindow: authFixedWindowLimiter,
			apiFixedWindow:  apiFixedWindowLimiter,
			tokenBucket:     apiTokenBuckerLimiter,
		},
	}

	router := app.mount()

	if err := app.run(router); err != nil {
		log.Panic("error starting the server")
	}
}
