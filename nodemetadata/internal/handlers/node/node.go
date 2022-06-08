package node

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/pkg/nodemetadata"
)

func Get(c *gin.Context) {
	nodesMetadata, err := nodemetadata.Get()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, nodesMetadata)
	}
}

func Put(c *gin.Context) {
	var input nodemetadata.NodeMetadata
	c.BindJSON(&input)
	err := nodemetadata.Add(&input)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}

func Delete(c *gin.Context) {
	id := c.Param("id")
	err := nodemetadata.Delete(id)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
