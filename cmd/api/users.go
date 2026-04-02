package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Petroviiic/finance_api_backend/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type UserPayload struct {
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Username  string `json:"username" validate:"required,max=100"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=3,max=72"`
}

func (app *Application) GetById(w http.ResponseWriter, r *http.Request) {

}

// RegisterUser godoc
// @Summary      Register a new user
// @Description  Creates a new user account.
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
	if err := Validate.Struct(data); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := &storage.User{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Username:  data.Username,
		Email:     data.Email,
		Role:      "viewer",
		IsActive:  true,
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

type LoginUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// Login godoc
// @Summary      User login
// @Description  Authenticates a user and returns a JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      LoginUserPayload  true  "Login credentials (username, password)"
// @Success      200      {string}  string       "JWT Token"
// @Failure      401      {object}  map[string]string "Invalid credentials"
// @Failure      500      {object}  map[string]string "Internal server error"
// @Router       /users/login [post]
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	var data LoginUserPayload
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
	if !user.IsActive {
		app.forbiddenResponse(w, r)
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

// GetUserInfo godoc
// @Summary      Get current user information
// @Description  Returns the profile information and role of the currently authenticated user based on the JWT token
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200      {object}  storage.User    "Successfully retrieved user profile"
// @Failure      401      {object}  error   "Unauthorized - Invalid or missing token"
// @Failure      500      {object}  error   "Internal server error"
// @Router       /users/me [get]
func (app *Application) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)

	if err := jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
}

type ChangeStatusRequest struct {
	IsActive bool `json:"is_active" validate:"required"`
}

// ChangeStatus godoc
// @Summary      Update user status
// @Description  Allows an admin to activate or deactivate a user account
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path      int     true  "User ID"
// @Param        payload body      ChangeStatusRequest  true  "New status (is_active: true/false)"
// @Success      200     {object}  map[string]string "user status updated successfully"
// @Failure      400     {object}  error             "Invalid request payload or ID"
// @Failure      401     {object}  error             "Unauthorized"
// @Failure      403     {object}  error             "Forbidden - Admin only"
// @Failure      500     {object}  error             "Internal server error"
// @Router       /users/{id}/status [patch]
func (app *Application) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input ChangeStatusRequest
	if err := readJson(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.storage.UserStorage.UpdateUserStatus(r.Context(), id, input.IsActive)
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, "user status updated successfully"); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
}

type ChangeRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=viewer analyst admin"`
}

// ChangeRole godoc
// @Summary      Update user role
// @Description  Allows an admin to change a user's access level (admin, analyst, viewer)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path      int     true  "User ID"
// @Param        payload body      ChangeRoleRequest  true  "New role name"
// @Success      200     {object}  map[string]string "user role updated successfully"
// @Failure      400     {object}  error             "Invalid role or ID"
// @Failure      401     {object}  error             "Unauthorized"
// @Failure      403     {object}  error             "Forbidden - Admin only"
// @Failure      500     {object}  error             "Internal server error"
// @Router       /users/{id}/role [patch]
func (app *Application) ChangeRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input ChangeRoleRequest
	if err := readJson(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.storage.UserStorage.UpdateUserRole(r.Context(), id, input.Role)
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
	if err := jsonResponse(w, http.StatusOK, "user role updated successfully"); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Permanently removes a user from the system by their ID. Only accessible by Admin.
// @Tags         users
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Produce      json
// @Success      200  {object} map[string]string  "user deleted successfully"
// @Failure      400  {object}  error   "Invalid user ID"
// @Failure      401  {object}  error   "Unauthorized"
// @Failure      403  {object}  error   "Forbidden - Admin only"
// @Failure      500  {object}  error   "Internal server error"
// @Router       /users/{id} [delete]
func (app *Application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.storage.UserStorage.DeleteUser(r.Context(), id)
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, "user deleted successfully"); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
}

// @Description  Returns a list of all registered users in the system. Only accessible by Admin.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object} map[string]string
// @Failure      401  {object}  error   "Unauthorized"
// @Failure      403  {object}  error   "Forbidden - Admin only"
// @Failure      500  {object}  error   "Internal server error"
// @Router       /users [get]
func (app *Application) GetAllUsers(w http.ResponseWriter, r *http.Request) {

	users, err := app.storage.UserStorage.GetAllUsers(r.Context())
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, users); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
}
