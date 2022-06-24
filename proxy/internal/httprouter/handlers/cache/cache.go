package cache

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/diegoximenes/distributed_cache/proxy/internal/httprouter/handlers"
	"github.com/diegoximenes/distributed_cache/proxy/internal/keypartition"
	"github.com/diegoximenes/distributed_cache/proxy/internal/logger"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/node"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/nodesmetadata"
	httpUtil "github.com/diegoximenes/distributed_cache/util/pkg/http"
)

func Get(
	nodeClient *node.NodeClient,
	nodesMetadataClient *nodesmetadata.NodesMetadataClient,
	keyPartitionStrategy keypartition.KeyPartitionStrategy,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		key := c.Param("key")

		nodeID, err := keyPartitionStrategy.GetNodeID(key)
		if err != nil {
			logger.Logger.Error(
				err.Error(),
				zap.String("handler", "cache.Get"),
			)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		nodeMetadata := nodesMetadataClient.NodesMetadata[nodeID]

		response, err := nodeClient.Get(nodeMetadata.Address, key)
		if err != nil {
			httpError, isHTTPError := err.(*httpUtil.HTTPError)
			if isHTTPError && (httpError.StatusCode == http.StatusNotFound) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			logger.Logger.Error(
				err.Error(),
				zap.String("handler", "cache.Get"),
				zap.String("key", key),
			)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func Put(
	nodeClient *node.NodeClient,
	nodesMetadataClient *nodesmetadata.NodesMetadataClient,
	keyPartitionStrategy keypartition.KeyPartitionStrategy,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input node.PutInput
		err := c.BindJSON(&input)
		if err != nil {
			apiError := handlers.APIError{
				Error: err.Error(),
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, apiError)
			return
		}

		nodeID, err := keyPartitionStrategy.GetNodeID(input.Key)
		if err != nil {
			logger.Logger.Error(
				err.Error(),
				zap.String("handler", "get.Put"),
			)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		nodeMetadata := nodesMetadataClient.NodesMetadata[nodeID]

		err = nodeClient.Put(nodeMetadata.Address, &input)
		if err != nil {
			logger.Logger.Error(
				err.Error(),
				zap.String("handler", "get.Put"),
				zap.String("key", input.Key),
			)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.AbortWithStatus(http.StatusOK)
	}
}
