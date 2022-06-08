package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/diegoximenes/distributed_key_value_cache/proxy/internal/config"
	"github.com/diegoximenes/distributed_key_value_cache/proxy/internal/handlers/cache"
	"github.com/diegoximenes/distributed_key_value_cache/proxy/internal/handlers/heartbeat"
	"github.com/diegoximenes/distributed_key_value_cache/proxy/internal/util/logger"
	"github.com/diegoximenes/distributed_key_value_cache/proxy/pkg/clients/configserver"
)

func main() {
	config.Read()

	err := logger.Init()
	if err != nil {
		panic(fmt.Sprintf("Error when setting logger: %v", err))
	}

	configServerClient, err := configserver.New()
	if err != nil {
		panic(fmt.Sprintf("Error when setting configServerClient: %v", err))
	}

	router := gin.Default()
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.GET("/cache/:key", cache.Get(configServerClient))
	router.PUT("/cache", cache.Put(configServerClient))
	router.Run()
}
