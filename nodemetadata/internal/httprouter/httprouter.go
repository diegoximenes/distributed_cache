package httprouter

import (
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/config"
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/handlers/node"
	"github.com/gin-gonic/gin"
)

func Set() {
	router := gin.Default()
	router.GET("/node", node.Get)
	router.PUT("/node", node.Put)
	router.DELETE("/node/:id", node.Delete)
	router.Run(config.Config.HTTPAddress)
}
