package main

import (
	"github.com/diegoximenes/distributed_key_value_cache/configserver/internal/handlers/node"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/node", node.Get)
	router.PUT("/node", node.Put)
	router.Run()
}
