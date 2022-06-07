package cache

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/diegoximenes/distributed_key_value_cache/proxy/internal/keypartition/rendezvoushashing"
	"github.com/diegoximenes/distributed_key_value_cache/proxy/pkg/clients/configserver"
	"github.com/diegoximenes/distributed_key_value_cache/proxy/pkg/clients/node"
	httpUtil "github.com/diegoximenes/distributed_key_value_cache/proxy/pkg/util/http"
)

func Get(configServerClient *configserver.ConfigServerClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		key := c.Param("key")

		nodeConfig := rendezvoushashing.GetNodeConfig(configServerClient.NodesConfig, key)

		response, err := node.Get(nodeConfig.Address, key)
		if err != nil {
			httpError, isHTTPError := err.(*httpUtil.HTTPError)
			if isHTTPError && (httpError.StatusCode == http.StatusNotFound) {
				c.AbortWithStatus(http.StatusNotFound)
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		} else {
			c.JSON(http.StatusOK, response)
		}
	}
}

func Put(configServerClient *configserver.ConfigServerClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input node.PutInput
		c.BindJSON(&input)

		nodeConfig := rendezvoushashing.GetNodeConfig(configServerClient.NodesConfig, input.Key)

		err := node.Put(nodeConfig.Address, &input)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			c.AbortWithStatus(http.StatusOK)
		}
	}
}
