package main

import (
	"encoding/json"
	"fmt"
	"github.com/sphireco/mantis"
	"net/http"
	"strconv"
	"time"
)

// HandleResponse Handles general responses via JSON
func HandleResponse(w http.ResponseWriter, val interface{}, status int) {
	// Set session token if available, as well as app name, request ID (for tracing) and version
	w.Header().Set("X-Session-Token", nextRequestID())
	w.Header().Set("X-Request-Id", nextRequestID())
	w.Header().Set("X-App-Name", fmt.Sprintf("%s %s", App.Name, App.Version))
	w.WriteHeader(status)

	// Output via a json encoder stream
	mantis.HandleError("HandleResponse Encode", json.NewEncoder(w).Encode(val))
}

// NotFoundServer Handles all not found with a simple error
func NotFoundServer(w http.ResponseWriter, req *http.Request) {
	HandleResponse(w, "404 Not Found", http.StatusNotFound)
}

// HomeServer Basic status and root handler
func HomeServer(w http.ResponseWriter, req *http.Request) {
	HandleResponse(w, "200 OK", http.StatusOK)
}

// GetStatus Returns the extended application status
func GetStatus(w http.ResponseWriter, req *http.Request) {
	HandleResponse(w, "200 OK", http.StatusOK)
}

// Teapot Easter Egg teapot 418 handler
func Teapot(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Teapot", "Chai")
	HandleResponse(w, "Are you a teapot?", http.StatusTeapot)
}

// GetRoutes Returns a listing of all routes
func GetRoutes(w http.ResponseWriter, req *http.Request) {
	HandleResponse(w, App.Router.Routes, http.StatusOK)
}

// nextRequestID Use unix timestamp
func nextRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
