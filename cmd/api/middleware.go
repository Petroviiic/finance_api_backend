package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Petroviiic/finance_api_backend/internal/ratelimiter"
	"github.com/Petroviiic/finance_api_backend/internal/storage"
	"github.com/golang-jwt/jwt/v5"
)

type userKey string

const userContextKey userKey = "user"

func (app *Application) TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}
		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header malformed"))
			return
		}

		token, err := app.authenticator.ValidateToken(parts[1])

		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx := r.Context()
		user, err := app.storage.UserStorage.GetById(ctx, userID)

		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}
		if !user.IsActive {
			app.customErrorJson(w, r, errors.New("account is not active"), http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(r *http.Request) *storage.User {
	return r.Context().Value(userContextKey).(*storage.User)
}

func (app *Application) RatelimiterMiddleware(limiter ratelimiter.Limiter, useUserID bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var key string
			if useUserID {
				id := GetUserFromContext(r).ID
				key = fmt.Sprintf("%d", id)
			} else {
				key = fmt.Sprintf("ip:%s", r.RemoteAddr)
			}

			if allow, retryAfter := limiter.Allow(key); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}

			next.ServeHTTP(w, r)
		})

	}
}

func (app *Application) RoleMiddleware(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requiredRoles) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			user := GetUserFromContext(r)

			for _, role := range requiredRoles {
				if user.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			app.forbiddenResponse(w, r)
		})

	}
}
