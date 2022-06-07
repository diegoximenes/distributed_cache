package node

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/diegoximenes/distributed_key_value_cache/configserver/pkg/nodeconfig"
)

func Get(c *gin.Context) {
	nodesConfig, err := nodeconfig.Get()
	if err != nil {
		c.Error(err)
	} else {
		c.JSON(http.StatusOK, nodesConfig)
	}
}

func Put(c *gin.Context) {
	var input nodeconfig.NodeConfig
	c.BindJSON(&input)
	err := nodeconfig.Add(&input)
	if err != nil {
		c.Error(err)
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
