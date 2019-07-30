package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sphireco/mantis"
	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"io/ioutil"
	"net/http"
	"time"
)

type Router struct {
	Routes      Routes
	router      *mux.Router
	cacheClient *cache.Client
}

type Routes struct {
	Route []Route `json:"routes"`
}

type Route struct {
	Name    string `json:"name"`
	Method  string `json:"method"`
	URI     string `json:"uri"`
	Handler string `json:"handler"`
}

// Routes Attach custom routes to mux
func (R *Router) Load() {
	R.newRouter()
	R.loadRoutes()
	R.startCache()

	for _, route := range R.Routes.Route {
		Logger.Write(fmt.Sprintf("Registering %s (%s %s)", route.Name, route.Method, route.URI))

		// Assign our routes to our handler methods
		// Could do this with reflect if an unknown number of routes
		// but chose this method for this assigment since it's just a few known routes
		var routeHandler func(w http.ResponseWriter, r *http.Request)
		switch route.Handler {
		case "GetStations":
			routeHandler = GetStations
		case "GetStationsInService":
			routeHandler = GetStationsInService
		case "GetStationsNotInService":
			routeHandler = GetStationsNotInService
		case "GetStationsMatchingString":
			routeHandler = GetStationsMatchingString
		case "GetIsBikeDockable":
			routeHandler = GetIsBikeDockable
		default:
			routeHandler = NotFoundServer
		}

		R.router.Methods(route.Method).
			Path(route.URI).
			Name(route.Name).
			Handler(R.cacheClient.Middleware(logRequest(http.HandlerFunc(routeHandler), route.Name)))
	}
}

func logRequest(next http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Logger.LogHTTPRequest(name, w, r)
		next.ServeHTTP(w, r)
	})
}

// defaultRoutes Creates our default routes
func (R *Router) newRouter() {
	R.router = mux.NewRouter().StrictSlash(false)

	// Register our default routes
	R.router.NotFoundHandler = http.HandlerFunc(NotFoundServer)
	R.router.HandleFunc("/", HomeServer)
	R.router.HandleFunc("/status", HomeServer)
	R.router.HandleFunc("/teapot", Teapot)
}

// loadRoutes Load all of our routes from routes.json
func (R *Router) loadRoutes() {
	routesJson, err := ioutil.ReadFile("routes.json")
	mantis.HandleFatalError(err)

	err = json.Unmarshal(routesJson, &R.Routes)
	mantis.HandleFatalError(err)
}

// startCache Start our in memory LRU cache
func (R *Router) startCache() {
	memoryCache, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	mantis.HandleFatalError(err)

	R.cacheClient, err = cache.NewClient(
		cache.ClientWithAdapter(memoryCache),
		cache.ClientWithTTL(App.Server.MemCacheTime*time.Minute),
		cache.ClientWithRefreshKey("opn"),
	)
	mantis.HandleFatalError(err)
}
