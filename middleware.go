package main

import (
	"net/http"
)

func (R *Router) registerMiddleWare() {
	R.middlewares = make(map[string]func(http.Handler) http.Handler)
	R.middlewares["adminOnly"] = adminOnly
}

// adminOnly Checks if a user is an admin
func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Logic here to determine if user is admin user
		next.ServeHTTP(w, r)
	})
}

// logRequest Middleware which logs each request
func logRequest(next http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Logger.LogHTTPRequest(name, w, r)
		next.ServeHTTP(w, r)
	})
}

// basicHeaders Apply our general headers
func basicHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		// Stop here if its Preflighted OPTIONS request
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}
