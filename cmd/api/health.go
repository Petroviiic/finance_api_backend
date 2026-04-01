package main

import (
	"log"
	"net/http"
)

// CreateGroup godoc
// @Summary      Health check
// @Description  Checks if the server is active.
// @Tags         utils
// @Success      200
// @Router       /health [get]
func (app *Application) GetHealth(w http.ResponseWriter, r *http.Request) {
	if err := jsonResponse(w, http.StatusOK, "ok"); err != nil {
		log.Panic("something went wrong")
	}
}
