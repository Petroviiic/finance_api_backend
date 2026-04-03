package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Petroviiic/finance_api_backend/internal/auth"
	"github.com/Petroviiic/finance_api_backend/internal/ratelimiter"
	"github.com/Petroviiic/finance_api_backend/internal/storage"
)

func newTestApplication(t *testing.T) *Application {
	t.Helper()

	storage := storage.NewMockStorage()
	auth := auth.NewMockJWTAuthenticator()
	return &Application{
		storage:       storage,
		authenticator: auth,
		rateLimiters: rateLimiters{
			apiFixedWindow:  ratelimiter.NewFixedWindowLimiter(1000, time.Second),
			authFixedWindow: ratelimiter.NewFixedWindowLimiter(1000, time.Second),
			tokenBucket:     ratelimiter.NewTokenBuckerRatelimiter(1000, 1000),
		},
	}

}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("expected the response code to be %d and we got %d", expected, actual)
	}
}
