package mux

import (
	"net"
	"time"
)

func Dial(network string, address string, timeout time.Duration, firstByte byte) (net.Conn, error) {
	netDialer := &net.Dialer{Timeout: timeout}
	conn, err := netDialer.Dial(network, address)
	if err != nil {
		return nil, err
	}

	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		conn.Close()
		return nil, err
	}
	if _, err := conn.Write([]byte{firstByte}); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
