package main

import (
	"log"
	"net/http"
)

func (app *Application) customErrorJson(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	log.Printf("error, method %s, path %s, error %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, statusCode, err.Error())
}
func (app *Application) internalServerErrorJson(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error, method %s, path %s, error %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *Application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	log.Printf("forbidden, method %s, path %s, error ", r.Method, r.URL.Path)
	writeJsonError(w, http.StatusForbidden, "forbidden")
}

func (app *Application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error, method %s, path %s, error %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, http.StatusBadRequest, err.Error())
}

func (app *Application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error, method %s, path %s, error %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, http.StatusNotFound, "not found")
}
func (app *Application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict error, method %s, path %s, error %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, http.StatusConflict, err.Error())
}
func (app *Application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("unauthorized error, method %s, path %s, error %s", r.Method, r.URL.Path, err.Error())
	writeJsonError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("unauthorized basic error, method %s, path %s, error %s", r.Method, r.URL.Path, err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	writeJsonError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	log.Printf("rate limit exceeded, method %s, path %s", r.Method, r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	writeJsonError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
