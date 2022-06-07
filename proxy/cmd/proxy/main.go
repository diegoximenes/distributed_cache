package main

import (
	"github.com/gin-gonic/gin"

	"github.com/diegoximenes/distributed_key_value_cache/proxy/internal/handlers/cache"
	"github.com/diegoximenes/distributed_key_value_cache/proxy/internal/handlers/heartbeat"
	"github.com/diegoximenes/distributed_key_value_cache/proxy/pkg/clients/configserver"
)

func main() {
	configServerClient, err := configserver.New()
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.GET("/cache/:key", cache.Get(configServerClient))
	router.PUT("/cache", cache.Put(configServerClient))
	router.Run()
}
