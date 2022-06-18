package cache

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/diegoximenes/distributed_cache/proxy/internal/handlers"
	"github.com/diegoximenes/distributed_cache/proxy/internal/keypartition/rendezvoushashing"
	"github.com/diegoximenes/distributed_cache/proxy/internal/util/logger"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/node"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/nodesmetadata"
	httpUtil "github.com/diegoximenes/distributed_cache/util/pkg/http"
)

func Get(
	nodeClient *node.NodeClient,
	nodesMetadataClient *nodesmetadata.NodesMetadataClient,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		key := c.Param("key")

		nodeMetadata := rendezvoushashing.GetNodeMetadata(
			&nodesMetadataClient.NodesMetadata,
			key,
		)
		if nodeMetadata == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		response, err := nodeClient.Get(nodeMetadata.Address, key)
		if err != nil {
			httpError, isHTTPError := err.(*httpUtil.HTTPError)
			if isHTTPError && (httpError.StatusCode == http.StatusNotFound) {
				c.AbortWithStatus(http.StatusNotFound)
			} else {
				logger.Logger.Error(
					err.Error(),
					zap.String("method", "get"),
					zap.String("key", key),
					zap.String("nodeMetadata.ID", nodeMetadata.ID),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		} else {
			c.JSON(http.StatusOK, response)
		}
	}
}

func Put(
	nodeClient *node.NodeClient,
	nodesMetadataClient *nodesmetadata.NodesMetadataClient,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input node.PutInput
		err := c.BindJSON(&input)
		if err != nil {
			apiError := handlers.APIError{
				Error: err.Error(),
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, apiError)
		}

		nodeMetadata := rendezvoushashing.GetNodeMetadata(
			&nodesMetadataClient.NodesMetadata,
			input.Key,
		)
		if nodeMetadata == nil {
			logger.Logger.Error(
				"Zero nodes available",
				zap.String("method", "put"),
			)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		err = nodeClient.Put(nodeMetadata.Address, &input)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			logger.Logger.Error(
				err.Error(),
				zap.String("method", "put"),
				zap.String("key", input.Key),
				zap.String("nodeMetadata.ID", nodeMetadata.ID),
			)
			return
		}

		c.AbortWithStatus(http.StatusOK)
	}
}
