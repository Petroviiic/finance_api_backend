package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Petroviiic/finance_api_backend/internal/storage"
	"github.com/golang-jwt/jwt/v5"
)

type UserPayload struct {
	Username              string `json:"username" validate:"required,max=100"`
	Email                 string `json:"email" validate:"omitempty,email,max=255"`
	Password              string `json:"password" validate:"required,min=3,max=72"`
	DeviceID              string `json:"device_id" validate:"required,max=255"`
	PushNotificationToken string `json:"push_notification_token"`
}

func (app *Application) GetById(w http.ResponseWriter, r *http.Request) {

}

// RegisterUser godoc
// @Summary      Register a new user
// @Description  Creates a new user account, hashes the password, and links the initial device with a push token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      UserPayload  true  "User registration data (email, username, password, device_id, push_token)"
// @Success      201      {nil}     nil          "User created successfully"
// @Failure      400      {object}  map[string]string "Invalid JSON or validation error"
// @Failure      400      {object}  map[string]string "User already exists"
// @Failure      500      {object}  map[string]string "Internal server error during hashing or database insert"
// @Router       /users/register [post]
func (app *Application) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var data UserPayload
	if err := readJson(w, r, &data); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if data.Email == "" {
		app.badRequestResponse(w, r, fmt.Errorf("email is required for registration"))
		return
	}
	if err := Validate.Struct(data); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := &storage.User{
		Email:    data.Email,
		Username: data.Username,
	}
	if err := user.Password.Set(data.Password); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	_, err := app.storage.UserStorage.RegisterUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, storage.ERROR_DUPLICATE_KEY_VALUE) {
			app.badRequestResponse(w, r, err)
			return
		}
		app.internalServerErrorJson(w, r, err)
		return
	}
	if err := jsonResponse(w, http.StatusCreated, nil); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
}

// Login godoc
// @Summary      User login
// @Description  Authenticates a user, updates their device token, and returns a JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      UserPayload  true  "Login credentials (username, password, device_id, push_token)"
// @Success      200      {string}  string       "JWT Token"
// @Failure      401      {object}  map[string]string "Invalid credentials"
// @Failure      500      {object}  map[string]string "Internal server error"
// @Router       /users/login [post]
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	var data UserPayload
	if err := readJson(w, r, &data); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(data); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	user, err := app.storage.UserStorage.GetByUsername(ctx, data.Username)
	if err != nil {
		app.unauthorizedErrorResponse(w, r, err)
		return
	}

	if !user.Password.ValidatePassword(data.Password) {
		app.unauthorizedErrorResponse(w, r, fmt.Errorf("unauthorized"))
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.expDate).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.iss,
		"aud": app.config.auth.iss,
	}
	token, err := app.authenticator.GenerateToken(claims)

	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
	if err := jsonResponse(w, http.StatusOK, token); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
}

// ValidateJWT godoc
// @Summary      Validate JWT
// @Description  Checks if jwt token is valid to skip login process
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Router       /users/validate_token [post]
// @Success      200      {string}  string       "Valid token"
// @Failure      401      {object}  map[string]string "Invalid token"
func (app *Application) ValidateJWTToken(w http.ResponseWriter, r *http.Request) {
	//TODO dodaj mzd ovdje da se updateuje last seen
}
