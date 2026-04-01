package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Petroviiic/finance_api_backend/internal/auth"
	"github.com/Petroviiic/finance_api_backend/internal/ratelimiter"
	"github.com/Petroviiic/finance_api_backend/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Application struct {
	config        Config
	db            *sql.DB
	storage       *storage.Storage
	authenticator auth.Authenticator
	rateLimiters  rateLimiters
}
type rateLimiters struct {
	apiFixedWindow  *ratelimiter.FixedWindowRateLimiter
	authFixedWindow *ratelimiter.FixedWindowRateLimiter
	tokenBucket     *ratelimiter.TokenBucketRatelimiter
}

type Config struct {
	addr        string
	isProdEnv   bool
	dbConfig    DBConfig
	auth        authConfig
	ratelimiter rateLimiterConfig
}

type rateLimiterConfig struct {
	authFixedWindow fixedWindowLimiterConfig
	apiFixedWindow  fixedWindowLimiterConfig
	tokenBucket     tokenBucketLimiterConfig
}
type tokenBucketLimiterConfig struct {
	limit           float64
	tokensPerMinute float64
}
type fixedWindowLimiterConfig struct {
	limit  int
	window time.Duration
}

type authConfig struct {
	secret  string
	expDate time.Duration
	iss     string
}
type DBConfig struct {
	maxIdleConns int
	maxOpenConns int
	maxIdleTime  string
	dbAddr       string
}

func (app *Application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000", // for local dev
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.GetHealth)

		r.Route("/users", func(r chi.Router) {
			r.Use(app.RatelimiterMiddleware(app.rateLimiters.authFixedWindow, false))

			r.Group(func(r chi.Router) {
				//r.Post("/get-by-id", app.GetById)
				r.Post("/register", app.RegisterUser)
				r.Post("/login", app.Login)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.TokenAuthMiddleware)
				r.Post("/validate_token", app.ValidateJWTToken)
			})
		})

	})

	return r
}
func (app *Application) run(router http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("starting server at %s", app.config.addr)

	return srv.ListenAndServe()
}
