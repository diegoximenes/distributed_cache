package main

import (
	"github.com/diegoximenes/distributed_cache/node/internal/config"
	"github.com/diegoximenes/distributed_cache/node/internal/httprouter"
	"github.com/diegoximenes/distributed_cache/node/internal/logger"
	cache "github.com/diegoximenes/distributed_cache/node/pkg/cache"
)

func main() {
	config.Read()

	err := logger.Init()
	if err != nil {
		panic(err)
	}
	defer logger.Logger.Sync()

	cache, err := cache.New(config.Config.CacheSize)
	if err != nil {
		panic(err)
	}

	httprouter.Set(cache)
}
