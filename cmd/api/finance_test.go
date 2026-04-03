package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestFinanceAccessControl(t *testing.T) {
	app := newTestApplication(t)
	mux := app.mount()

	type testCase struct {
		name               string
		method             string
		url                string
		userid             int64
		userRole           string // "admin", "analyst", "viewer", ili "unauthenticated"
		expectedStatusCode int
	}

	tests := []testCase{
		// create record - admin only
		{
			name:               "Admin can create record",
			method:             "POST",
			url:                "/v1/finance/create_record",
			userRole:           "admin",
			userid:             1,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Analyst cannot create record",
			method:             "POST",
			url:                "/v1/finance/create_record",
			userRole:           "analyst",
			userid:             2,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "Viewer cannot create record",
			method:             "POST",
			url:                "/v1/finance/create_record",
			userRole:           "viewer",
			userid:             3,
			expectedStatusCode: http.StatusForbidden,
		},

		// trends - admin and analyst
		{
			name:               "Analyst can view trends",
			method:             "GET",
			url:                "/v1/finance/trends?months_back=6",
			userRole:           "analyst",
			userid:             2,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Viewer cannot view trends",
			method:             "GET",
			url:                "/v1/finance/trends?months_back=6",
			userRole:           "viewer",
			userid:             3,
			expectedStatusCode: http.StatusForbidden,
		},

		// summary - everyone
		{
			name:               "Viewer can view summary",
			method:             "GET",
			url:                "/v1/finance/summary",
			userRole:           "viewer",
			userid:             3,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthenticated user cannot view summary",
			method:             "GET",
			url:                "/v1/finance/summary",
			userRole:           "unauthenticated",
			userid:             5,
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]interface{}{
				"amount":         1,
				"target_user_id": 1,
				"type":           "expense",
				"category":       "test",
				"entry_date":     "2006-04-05",
				"description":    "desc",
			}
			jsonBody, _ := json.Marshal(body)

			req, _ := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			if tc.userRole != "unauthenticated" {
				token, _ := app.authenticator.GenerateToken(jwt.MapClaims{
					"sub": tc.userid,
					"exp": time.Now().Add(time.Hour).Unix(),
				})
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp := executeRequest(req, mux)

			checkResponseCode(t, tc.expectedStatusCode, resp.Code)
		})
	}
}
