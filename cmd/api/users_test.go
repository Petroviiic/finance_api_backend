package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestUsersSystem(t *testing.T) {
	app := newTestApplication(t)
	mux := app.mount()

	type testCase struct {
		name               string
		method             string
		url                string
		userRole           string
		body               any
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:   "Register: Success",
			method: "POST",
			url:    "/v1/users/register",
			body: map[string]string{
				"first_name": "name",
				"last_name":  "last_name",
				"username":   "new_user",
				"password":   "password123",
				"email":      "new@example.com",
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:   "Login: Success",
			method: "POST",
			url:    "/v1/users/login",
			body: map[string]string{
				"username": "admin_user",
				"password": "password123",
			},
			expectedStatusCode: http.StatusOK,
		},

		// --- AUTHENTICATED (ME) ---
		{
			name:               "Get Me: Success as Viewer",
			method:             "GET",
			url:                "/v1/users/me",
			userRole:           "viewer",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Get Me: Failure (No Token)",
			method:             "GET",
			url:                "/v1/users/me",
			userRole:           "unauthenticated",
			expectedStatusCode: http.StatusUnauthorized,
		},

		// --- ADMIN ONLY ROUTES ---
		{
			name:               "Admin: List all users - Success",
			method:             "GET",
			url:                "/v1/users/",
			userRole:           "admin",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Viewer: List all users - Forbidden",
			method:             "GET",
			url:                "/v1/users/",
			userRole:           "viewer",
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "Admin: Change User Status - Success",
			method:             "PATCH",
			url:                "/v1/users/2/status",
			userRole:           "admin",
			body:               map[string]interface{}{"is_active": false},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Viewer: Try to change role - Forbidden",
			method:             "PATCH",
			url:                "/v1/users/3/role",
			userRole:           "viewer",
			body:               map[string]string{"role": "admin"},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "Admin: Delete User - Success",
			method:             "DELETE",
			url:                "/v1/users/2",
			userRole:           "admin",
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.body != nil {
				jsonBody, _ := json.Marshal(tc.body)
				bodyReader = bytes.NewBuffer(jsonBody)
			}

			req, _ := http.NewRequest(tc.method, tc.url, bodyReader)
			req.Header.Set("Content-Type", "application/json")

			if tc.userRole != "" && tc.userRole != "unauthenticated" {
				userID := int64(3)
				if tc.userRole == "admin" {
					userID = 1
				}

				claims := jwt.MapClaims{
					"sub": userID,
					"exp": time.Now().Add(time.Hour).Unix(),
				}
				token, _ := app.authenticator.GenerateToken(claims)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp := executeRequest(req, mux)

			checkResponseCode(t, tc.expectedStatusCode, resp.Code)
		})
	}
}
