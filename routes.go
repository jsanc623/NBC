package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/victorspringer/http-cache"
	"net/http"
)

type handler func(http.ResponseWriter, *http.Request)

// Router Define our core Router struct
type Router struct {
	Routes      []Route `json:"routes"`
	router      *mux.Router
	httpCache   *cache.Client
	middlewares map[string]func(http.Handler) http.Handler
}

// Route Define a route
type Route struct {
	Name       string   `json:"name"`
	Method     string   `json:"method"`
	URI        string   `json:"uri"`
	Middleware []string `json:"middleware"`
	handler    func(http.ResponseWriter, *http.Request)
}

// newRoutes Load our default routes
func (R *Router) newRoutes() {
	R.new("GetStations", "GET", "/stations", GetStations, []string{})
	R.new("GetStationsInService", "GET", "/stations/in-service", GetStationsInService, []string{})
	R.new("GetStationsNotInService", "GET", "/stations/not-in-service", GetStationsNotInService, []string{})
	R.new("GetStationsMatchingString", "GET", "/stations/{search}", GetStationsMatchingString, []string{})
	R.new("GetIsBikeDockable", "GET", "/dockable/{stationId}/{bikesToReturn}", GetIsBikeDockable, []string{})
}

// Load Create a new router and attach our default and custom routes
func (R *Router) Load() {
	R.router = mux.NewRouter().StrictSlash(false)
	R.startCache()
	R.registerMiddleWare()

	R.router.NotFoundHandler = http.HandlerFunc(NotFoundServer)
	R.new("HomeServer", "GET", "/", HomeServer, []string{})
	R.new("Status", "GET", "/status", GetStatus, []string{})
	R.new("Teapot", "GET", "/teapot", Teapot, []string{})
	R.new("GetRoutes", "GET", "/routes", GetRoutes, []string{})

	R.newRoutes()

	for _, route := range R.Routes {
		Logger.Write(fmt.Sprintf("Activating %s (%s %s)", route.Name, route.Method, route.URI))
		R.addRoute(route)
	}
}

// new Append a new route to our routes
func (R *Router) new(name string, method string, uri string, handler handler, middleware []string) {
	Logger.Write(fmt.Sprintf("Registering %s (%s %s)", name, method, uri))
	R.Routes = append(R.Routes, Route{
		Name:       name,
		Method:     method,
		URI:        uri,
		handler:    handler,
		Middleware: middleware,
	})
}

// addRoute Add a route to our router
func (R *Router) addRoute(route Route) {
	// Apply our two forced middlewares
	handler := logRequest(http.HandlerFunc(route.handler), route.Name)
	handler = basicHeaders(handler)

	// Apply all of our other middlewares specific to this route
	if len(route.Middleware) > 0 {
		for _, middleware := range route.Middleware {
			handler = R.middlewares[middleware](handler)
		}
	}

	handler = R.httpCache.Middleware(handler)
	R.router.Methods(route.Method).Path(route.URI).Name(route.Name).Handler(handler)
}
