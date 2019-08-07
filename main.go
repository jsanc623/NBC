package main

import (
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/go-redis/redis"
	"github.com/sphireco/mantis"
	"github.com/subosito/gotenv"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Application Define the application
type Application struct {
	Name     string             `json:"name"`
	ID       string             `json:"id"`
	Version  string             `json:"version"`
	Log      string             `json:"log_location"`
	Runtime  string             `json:"runtime"`
	Server   Server             `json:"server"`
	Token    string             `json:"token"`
	Router   Router             `json:"routes"`
	Cache    *bigcache.BigCache `json:"cache"`
	Redis    *redis.Client      `json:"redis"`
	Database string             `json:"database"`
	Emailer  string             `json:"emailer"`
}

// Server Defines our core Server
type Server struct {
	Address      string        `json:"address"`
	Port         string        `json:"port"`
	WriteTimeout time.Duration `json:"write_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	MemCacheTime time.Duration `json:"mem_cache_time"`
}

// App The Core Application Definitions
var App Application

// Logger The main logging interface
var Logger mantis.Log

func init() {
	err := gotenv.Load()
	if err != nil {
		panic(err)
	}

	writeTimeout, err1 := strconv.Atoi(os.Getenv("SRV_WRITE_TIMEOUT"))
	readTimeout, err2 := strconv.Atoi(os.Getenv("SRV_READ_TIMEOUT"))
	memCacheTime, err3 := strconv.Atoi(os.Getenv("SRV_MEMCACHE_TIME_MINUTES"))

	// Set default values if atoi fails
	if err1 != nil || err2 != nil || err3 != nil {
		writeTimeout = 30
		readTimeout = 30
		memCacheTime = 30
	}

	App = Application{
		Name:    os.Getenv("APP_NAME"),
		ID:      os.Getenv("APP_ID"),
		Version: os.Getenv("APP_VERSION"),
		Log:     os.Getenv("LOG_LOCATION"),
		Server: Server{
			Address:      os.Getenv("SRV_ADDRESS"),
			Port:         os.Getenv("SRV_PORT"),
			WriteTimeout: time.Duration(writeTimeout),
			ReadTimeout:  time.Duration(readTimeout),
			MemCacheTime: time.Duration(memCacheTime),
		},
		Runtime: time.Now().UTC().Format(time.RFC3339),
	}

	if len(os.Getenv("REDIS_ADDRESS")) > 0 {
		App.Redis = setupRedis()
	}

	config := bigcache.Config{
		Shards:             1024,
		LifeWindow:         10 * time.Minute,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       1024,
		Verbose:            true,
		HardMaxCacheSize:   64,
	}
	App.Cache, err = bigcache.NewBigCache(config)
	mantis.HandleFatalError(err)

	Logger.NewLog(App.Log)
	mantis.SetErrorLog(Logger)

	Logger.Write(fmt.Sprintf("Initializing %s %s @ %s", App.Name, App.Version, App.Runtime))
	Logger.Write(App.ID + "\n")
}

func main() {
	App.Router.Load()
	srv := &http.Server{
		Handler:      App.Router.router,
		Addr:         fmt.Sprintf("%s:%s", App.Server.Address, App.Server.Port),
		WriteTimeout: App.Server.WriteTimeout * time.Second,
		ReadTimeout:  App.Server.ReadTimeout * time.Second,
	}

	mantis.HandleError("ListenAndServe", srv.ListenAndServe())
}

func setupRedis() *redis.Client {
	readTimeout, err := strconv.Atoi(os.Getenv("REDIS_READ_TIMEOUT"))
	mantis.HandleFatalError(err)
	writeTimeout, err := strconv.Atoi(os.Getenv("REDIS_WRITE_TIMEOUT"))
	mantis.HandleFatalError(err)
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	mantis.HandleFatalError(err)
	maxRetries, err := strconv.Atoi(os.Getenv("REDIS_MAX_RETRIES"))
	mantis.HandleFatalError(err)
	minRetryBackoff, err := strconv.Atoi(os.Getenv("REDIS_MIN_RETRY_BACKOFF"))
	mantis.HandleFatalError(err)
	maxRetryBackoff, err := strconv.Atoi(os.Getenv("REDIS_MAX_RETRY_BACKOFF"))
	mantis.HandleFatalError(err)
	dialTimeout, err := strconv.Atoi(os.Getenv("REDIS_DIAL_TIMEOUT"))
	mantis.HandleFatalError(err)
	poolSize, err := strconv.Atoi(os.Getenv("REDIS_POOL_SIZE"))
	mantis.HandleFatalError(err)
	minIdleConns, err := strconv.Atoi(os.Getenv("REDIS_MIN_IDLE_CONNS"))
	mantis.HandleFatalError(err)

	client := redis.NewClient(&redis.Options{
		Addr:            os.Getenv("REDIS_ADDRESS"),
		Password:        os.Getenv("REDIS_PASSWORD"),
		DB:              db,
		MaxRetries:      maxRetries,
		MinRetryBackoff: time.Duration(minRetryBackoff),
		MaxRetryBackoff: time.Duration(maxRetryBackoff),
		DialTimeout:     time.Duration(dialTimeout),
		ReadTimeout:     time.Duration(readTimeout) * time.Second,
		WriteTimeout:    time.Duration(writeTimeout) * time.Second,
		PoolSize:        poolSize,
		MinIdleConns:    minIdleConns,
	})

	pong, err := client.Ping().Result()
	mantis.HandleFatalError(err)
	fmt.Println(pong) // Output: PONG

	return client
}
