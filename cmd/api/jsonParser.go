package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
	_ = Validate.RegisterValidation("validduration", func(fl validator.FieldLevel) bool {
		durationStr := fl.Field().String()
		_, err := time.ParseDuration(durationStr)
		return err == nil
	})
}

func readJson(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func writeJson(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func writeJsonError(w http.ResponseWriter, status int, message string) error {

	type jsonError struct {
		Error string `json:"error"`
	}

	return writeJson(w, status, jsonError{Error: message})
}

func jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}
	return writeJson(w, status, envelope{Data: data})
}
