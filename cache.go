package main

import (
	"github.com/sphireco/mantis"
	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"time"
)

// startCache Start our in memory LRU cache
func (R *Router) startCache() {
	memoryCache, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	mantis.HandleFatalError(err)

	R.httpCache, err = cache.NewClient(
		cache.ClientWithAdapter(memoryCache),
		cache.ClientWithTTL(App.Server.MemCacheTime*time.Minute),
		cache.ClientWithRefreshKey("opn"),
	)
	mantis.HandleFatalError(err)
}
