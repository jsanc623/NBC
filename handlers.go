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
	w.WriteHeader(status)

	// Set session token if available, as well as app name, request ID (for tracing) and version
	w.Header().Set("X-Session-Token", "")
	w.Header().Set("X-Request-Id", nextRequestID())
	w.Header().Set("X-App-Name", fmt.Sprintf("%s %s", App.Name, App.Version))

	mantis.HandleError("HandleResponse Encode", json.NewEncoder(w).Encode(val))
}

// nextRequestID Use unix timestamp
func nextRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// NotFoundServer Handles all not found with a simple error
func NotFoundServer(w http.ResponseWriter, req *http.Request) {
	Logger.LogHTTPRequest("NotFoundServer", w, req)
	HandleResponse(w, "404 Not Found", http.StatusNotFound)
}

// HomeServer Basic status and root handler
func HomeServer(w http.ResponseWriter, req *http.Request) {
	Logger.LogHTTPRequest("Home", w, req)
	HandleResponse(w, "200 OK", http.StatusOK)
}

func Teapot(w http.ResponseWriter, req *http.Request) {
	Logger.LogHTTPRequest("Teapot", w, req)
	w.Header().Set("X-Teapot", "Chai")
	HandleResponse(w, "Are you a teapot?", http.StatusTeapot)
}
