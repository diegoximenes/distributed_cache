package demux

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/diegoximenes/distributed_cache/nodemetadata/pkg/net/connection/listener"
)

type Demux struct {
	inListener   net.Listener
	outListeners map[byte]*listener.Listener
}

func New(address string) (*Demux, error) {
	inListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	demux := Demux{
		inListener:   inListener,
		outListeners: make(map[byte]*listener.Listener),
	}
	go demux.serve()

	return &demux, nil
}

func (demux *Demux) serve() error {
	for {
		conn, err := demux.inListener.Accept()
		netOpError, isNetOpError := err.(*net.OpError)
		if isNetOpError && netOpError.Temporary() {
			continue
		}

		if err != nil {
			panic(err)
		}

		go demux.handleConn(conn)
	}
}

func (demux *Demux) handleConn(conn net.Conn) {
	// set a read deadline so connections with no data timeout
	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		conn.Close()
		return
	}

	// read first byte from connection to determine out listener
	var connType [1]byte
	if _, err := io.ReadFull(conn, connType[:]); err != nil {
		conn.Close()
		return
	}

	// reset read deadline and let the out listener handle that
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		conn.Close()
		return
	}

	// retrieve out listener based on first byte
	outListener, outListenerExists := demux.outListeners[connType[0]]
	if !outListenerExists {
		conn.Close()
		return
	}

	// send connection to out listener, which will be responsible for closing the connection
	outListener.Handle(conn)
}

func (demux *Demux) RegisterOutListener(firstByte byte, outListener *listener.Listener) error {
	if _, outListenerExists := demux.outListeners[firstByte]; outListenerExists {
		return errors.New(fmt.Sprintf("firstByte %v already has an outListener registered", firstByte))
	}

	demux.outListeners[firstByte] = outListener
	return nil
}
