package metadata

import (
	"net"
	"net/http"

	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/config"
	"github.com/diegoximenes/distributed_cache/nodemetadata/pkg/net/connection/demux"
	"github.com/diegoximenes/distributed_cache/nodemetadata/pkg/net/connection/listener"
	"github.com/gin-gonic/gin"
)

func SetServer(demux *demux.Demux, tcpAddr *net.TCPAddr, firstByte byte) {
	raftNodeMetadataListener := listener.New(tcpAddr)
	demux.RegisterOutListener(firstByte, raftNodeMetadataListener)

	router := gin.Default()
	router.GET(HTTPPath, func(c *gin.Context) {
		response := Response{
			ApplicationAddress: config.Config.ApplicationAddress,
		}
		c.JSON(http.StatusOK, response)
	})

	go func() {
		err := http.Serve(raftNodeMetadataListener, router)
		if err != nil {
			panic(err)
		}
	}()
}
