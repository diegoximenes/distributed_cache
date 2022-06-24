package streamlayer

import (
	"net"
	"time"

	"github.com/diegoximenes/distributed_cache/nodesmetadata/pkg/net/connection/listener"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/pkg/net/connection/mux"
	"github.com/hashicorp/raft"
)

type StreamLayer struct {
	inListener    *listener.Listener
	dialFirstByte byte
}

func New(inListener *listener.Listener, dialFirstByte byte) *StreamLayer {
	return &StreamLayer{
		inListener:    inListener,
		dialFirstByte: dialFirstByte,
	}
}

func (streamLayer *StreamLayer) Dial(
	address raft.ServerAddress,
	timeout time.Duration,
) (net.Conn, error) {
	return mux.Dial("tcp", string(address), timeout, streamLayer.dialFirstByte)
}

func (streamLayer *StreamLayer) Accept() (net.Conn, error) {
	return streamLayer.inListener.Accept()
}

func (streamLayer *StreamLayer) Close() error {
	return streamLayer.inListener.Close()
}

func (streamLayer *StreamLayer) Addr() net.Addr {
	return streamLayer.inListener.Addr()
}
